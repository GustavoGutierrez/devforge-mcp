#!/usr/bin/env bash
# scripts/seed.sh — Idempotent database seed script for dev-forge-mcp
#
# Usage:
#   ./scripts/seed.sh [OPTIONS]
#
# Options:
#   --db PATH            Path to the libSQL database file (default: ./db/ui_patterns.db)
#   --ollama-url URL     Ollama base URL (default: http://localhost:11434)
#   --ollama-model MODEL Embedding model name (default: nomic-embed-text)
#   --seeds-only         Only apply SQL seeds (skip schema init and embeddings)
#   --embeddings-only    Only generate embeddings (skip schema init and seeds)
#   --no-embeddings      Skip embedding generation
#   --help               Show this message
#
# The script is idempotent: safe to run multiple times.
# Seeds use INSERT OR IGNORE so re-running never duplicates rows.

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

DB_PATH="${PROJECT_DIR}/dist/devforge.db"
OLLAMA_URL="http://localhost:11434"
OLLAMA_MODEL="nomic-embed-text"
SEEDS_DIR="${PROJECT_DIR}/db/seeds"

MODE="all"   # all | seeds-only | embeddings-only

# ── Argument parsing ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
    case "$1" in
        --db)           DB_PATH="$2";      shift 2 ;;
        --ollama-url)   OLLAMA_URL="$2";   shift 2 ;;
        --ollama-model) OLLAMA_MODEL="$2"; shift 2 ;;
        --seeds-only)   MODE="seeds";      shift   ;;
        --embeddings-only) MODE="embeddings"; shift ;;
        --no-embeddings) MODE="seeds";     shift   ;;
        --help)
            grep '^#' "$0" | grep -v '^#!/' | sed 's/^# \?//'
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# ── Colors ────────────────────────────────────────────────────────────────────
if [ -t 1 ]; then
    GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; RESET='\033[0m'
else
    GREEN=''; YELLOW=''; RED=''; RESET=''
fi

info()    { echo -e "${GREEN}[seed]${RESET} $*"; }
warn()    { echo -e "${YELLOW}[seed]${RESET} $*"; }
error()   { echo -e "${RED}[seed]${RESET} $*" >&2; }
section() { echo -e "\n${GREEN}━━ $* ━━${RESET}"; }

# ── Prerequisites ─────────────────────────────────────────────────────────────
check_prereqs() {
    section "Checking prerequisites"

    # Go is required to run the schema initializer and embedding helper
    if ! command -v go &>/dev/null; then
        error "go not found — install Go 1.24+ and ensure CGO_ENABLED=1 builds work"
        exit 1
    fi
    info "go: $(go version | awk '{print $3}')"

    # curl is needed for Ollama HTTP calls
    if ! command -v curl &>/dev/null; then
        warn "curl not found — embedding generation will be skipped"
        OLLAMA_AVAILABLE=false
    else
        check_ollama
    fi
}

check_ollama() {
    if curl -sf --max-time 2 "${OLLAMA_URL}/api/tags" >/dev/null 2>&1; then
        info "Ollama: available at ${OLLAMA_URL}"
        OLLAMA_AVAILABLE=true
    else
        warn "Ollama not reachable at ${OLLAMA_URL} — embedding generation will be skipped"
        OLLAMA_AVAILABLE=false
    fi
}

# ── Schema initialization ─────────────────────────────────────────────────────
# The schema is embedded in Go code (internal/db/schema.go) and applied by db.Open().
# We invoke scripts/init_db_runner.go directly — it calls db.Open() / RunMigrations()
# and exits immediately.  This is reliable and does not require starting the MCP server.

init_schema() {
    section "Initializing schema"

    DB_DIR="$(dirname "${DB_PATH}")"
    mkdir -p "${DB_DIR}"

    info "Running schema migrations via init_db_runner..."
    (
        cd "${PROJECT_DIR}"
        CGO_ENABLED=1 go run ./scripts/init_db_runner -db "${DB_PATH}"
    )

    if [ -f "${DB_PATH}" ]; then
        info "Schema applied — database ready at ${DB_PATH}"
    else
        error "DB file not found after init — something went wrong."
        exit 1
    fi
}

# ── SQL seed application ───────────────────────────────────────────────────────
# We use scripts/seed_runner.go (which uses the go-libsql driver) instead of the
# sqlite3 CLI because the schema has libSQL-specific types (F32_BLOB, libsql_vector_idx).

apply_seeds() {
    section "Applying SQL seeds"

    if [ ! -d "${SEEDS_DIR}" ]; then
        warn "Seeds directory not found: ${SEEDS_DIR}"
        return 0
    fi

    SEED_FILES=()
    while IFS= read -r -d '' f; do
        SEED_FILES+=("$f")
    done < <(find "${SEEDS_DIR}" -name '*.sql' -print0 | sort -z)

    if [ ${#SEED_FILES[@]} -eq 0 ]; then
        warn "No seed files found in ${SEEDS_DIR}"
        return 0
    fi

    info "Found ${#SEED_FILES[@]} seed file(s):"
    for f in "${SEED_FILES[@]}"; do
        echo "  $(basename "$f")"
    done

    # Ensure DB exists before seeding
    if [ ! -f "${DB_PATH}" ]; then
        warn "Database file ${DB_PATH} does not exist — schema must be initialized first"
        warn "Run 'make db-init' or start the server once with 'make run'"
        return 1
    fi

    # Build the -sql flag list
    SQL_ARGS=()
    for seed_file in "${SEED_FILES[@]}"; do
        SQL_ARGS+=("-sql" "${seed_file}")
    done

    (
        cd "${PROJECT_DIR}"
        CGO_ENABLED=1 go run ./scripts/seed_runner -db "${DB_PATH}" "${SQL_ARGS[@]}"
    )
}

# ── Embedding generation ───────────────────────────────────────────────────────
# For each pattern/architecture row with embedding IS NULL, call Ollama's
# /api/embeddings endpoint and store the result as a little-endian F32_BLOB.
# This is a standalone Go program that matches the encoding used by embed.go.

generate_embeddings() {
    section "Generating embeddings"

    if [ "${OLLAMA_AVAILABLE}" != "true" ]; then
        warn "Ollama not available — skipping embedding generation"
        warn "Start Ollama and run: make db-embeddings"
        return 0
    fi

    if [ ! -f "${DB_PATH}" ]; then
        warn "Database not found — skipping embedding generation"
        return 0
    fi

    info "Generating embeddings via Ollama (${OLLAMA_MODEL} at ${OLLAMA_URL})..."

    TMP_EMBEDDER_DIR="$(mktemp -d /tmp/dev-forge-embedder-XXXXXX)"
    trap "rm -rf '${TMP_EMBEDDER_DIR}'" EXIT

    cat > "${TMP_EMBEDDER_DIR}/main.go" <<'GOEOF'
package main

import (
	"context"
	"database/sql"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"

	ollamaapi "github.com/ollama/ollama/api"
	_ "github.com/tursodatabase/go-libsql"
)

func main() {
	dbPath     := flag.String("db", "", "Path to database")
	ollamaURL  := flag.String("ollama-url", "http://localhost:11434", "Ollama base URL")
	model      := flag.String("model", "nomic-embed-text", "Embedding model")
	flag.Parse()

	if *dbPath == "" {
		log.Fatal("usage: embedder -db <path> [-ollama-url URL] [-model MODEL]")
	}

	db, err := sql.Open("libsql", "file:"+*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	client := newOllamaClient(*ollamaURL)
	total := 0

	tables := []struct {
		table string
		expr  string
	}{
		{"patterns", "name || ' ' || COALESCE(description,'') || ' ' || COALESCE(tags,'')"},
		{"architectures", "name || ' ' || COALESCE(description,'') || ' ' || COALESCE(tags,'')"},
	}

	for _, t := range tables {
		rows, err := db.Query(
			"SELECT id, " + t.expr + " AS text FROM " + t.table + " WHERE embedding IS NULL LIMIT 500",
		)
		if err != nil {
			log.Printf("query %s: %v", t.table, err)
			continue
		}

		type row struct{ id, text string }
		var pending []row
		for rows.Next() {
			var r row
			if rows.Scan(&r.id, &r.text) == nil {
				pending = append(pending, r)
			}
		}
		rows.Close()

		fmt.Printf("  %s: %d row(s) need embeddings\n", t.table, len(pending))
		for _, r := range pending {
			vec, err := embed(client, *model, r.text)
			if err != nil || vec == nil {
				log.Printf("  embed %s/%s: %v", t.table, r.id, err)
				continue
			}
			blob := encodeF32(vec)
			if _, err := db.Exec("UPDATE "+t.table+" SET embedding = ? WHERE id = ?", blob, r.id); err != nil {
				log.Printf("  update %s/%s: %v", t.table, r.id, err)
				continue
			}
			total++
			fmt.Printf("  embedded %s/%s (%d dims)\n", t.table, r.id, len(vec))
		}
	}
	fmt.Printf("Done — %d embedding(s) generated\n", total)
}

func newOllamaClient(rawURL string) *ollamaapi.Client {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("parse ollama url: %v", err)
	}
	return ollamaapi.NewClient(u, &http.Client{Timeout: 30 * time.Second})
}

func embed(client *ollamaapi.Client, model, text string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &ollamaapi.EmbeddingRequest{Model: model, Prompt: text}
	resp, err := client.Embeddings(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Embedding) == 0 {
		return nil, nil
	}
	result := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		result[i] = float32(v)
	}
	return result, nil
}

// encodeF32 converts float32 slice to little-endian bytes (F32_BLOB format).
func encodeF32(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, f := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}
GOEOF

    cp "${PROJECT_DIR}/go.mod" "${TMP_EMBEDDER_DIR}/go.mod"
    cp "${PROJECT_DIR}/go.sum" "${TMP_EMBEDDER_DIR}/go.sum" 2>/dev/null || true

    EMBEDDER_BIN="${TMP_EMBEDDER_DIR}/embedder"
    (
        cd "${TMP_EMBEDDER_DIR}"
        CGO_ENABLED=1 go build -o "${EMBEDDER_BIN}" . 2>&1
    )

    "${EMBEDDER_BIN}" \
        -db "${DB_PATH}" \
        -ollama-url "${OLLAMA_URL}" \
        -model "${OLLAMA_MODEL}"

    rm -rf "${TMP_EMBEDDER_DIR}"
    trap - EXIT
}

# ── Count helper ──────────────────────────────────────────────────────────────
report_counts() {
    section "Summary"

    if [ ! -f "${DB_PATH}" ]; then
        warn "Database not found at ${DB_PATH}"
        return
    fi

    # Use a quick Go one-liner to query counts via libsql driver
    COUNT_OUTPUT="$(
        cd "${PROJECT_DIR}"
        CGO_ENABLED=1 go run - "${DB_PATH}" <<'GOCOUNT'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    _ "github.com/tursodatabase/go-libsql"
)

func main() {
    db, err := sql.Open("libsql", "file:"+os.Args[1])
    if err != nil { log.Fatal(err) }
    defer db.Close()

    tables := []string{"patterns", "architectures", "palettes", "tokens"}
    for _, t := range tables {
        var total, withEmbed int
        db.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&total)
        // Only patterns/architectures have embedding column
        if t == "patterns" || t == "architectures" {
            db.QueryRow("SELECT COUNT(*) FROM " + t + " WHERE embedding IS NOT NULL").Scan(&withEmbed)
            fmt.Printf("  %-15s %d rows, %d with embeddings\n", t+":", total, withEmbed)
        } else {
            fmt.Printf("  %-15s %d rows\n", t+":", total)
        }
    }
}
GOCOUNT
    2>/dev/null || echo "  (could not query DB)"
    )"

    info "Database: ${DB_PATH}"
    echo "${COUNT_OUTPUT}"
}

# ── Main ──────────────────────────────────────────────────────────────────────
main() {
    echo ""
    info "dev-forge-mcp seed script"
    info "DB: ${DB_PATH}"
    info "Ollama: ${OLLAMA_URL} (model: ${OLLAMA_MODEL})"
    info "Mode: ${MODE}"

    OLLAMA_AVAILABLE=false

    case "${MODE}" in
        all)
            check_prereqs
            init_schema
            apply_seeds
            generate_embeddings
            report_counts
            ;;
        seeds)
            check_prereqs
            init_schema
            apply_seeds
            report_counts
            ;;
        embeddings)
            check_prereqs
            check_ollama
            generate_embeddings
            report_counts
            ;;
    esac

    echo ""
    info "Done."
}

main "$@"
