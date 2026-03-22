# Makefile — dev-forge-mcp
# Requires: CGO_ENABLED=1 (libsql / go-sqlite3)
# Usage: make help

# ── Variables ──────────────────────────────────────────────────────────────────
BINARY_MCP   := dev-forge-mcp
BINARY_TUI   := dev-forge
DIST_DIR     := dist
BIN_DIR      := bin

# DB path used by the app at runtime (relative to project root).
# The canonical distribution DB lives inside dist/ so the binary and DB ship together.
DB_PATH      := $(DIST_DIR)/dev-forge.db

# Seeds directory
SEEDS_DIR    := db/seeds

# Ollama settings — override on the command line:
#   make db-embeddings OLLAMA_URL=http://myhost:11434 OLLAMA_MODEL=mxbai-embed-large
OLLAMA_URL   := http://localhost:11434
OLLAMA_MODEL := nomic-embed-text

# Install destination
INSTALL_DIR  := $(HOME)/.local/bin

# Go build flags
GO_BUILD     := CGO_ENABLED=1 go build
GO_RUN       := CGO_ENABLED=1 go run
GO_TEST      := CGO_ENABLED=1 go test

# Helper scripts (Go programs, run directly via go run — no pre-compilation needed)
INIT_RUNNER  := ./scripts/init_db_runner
SEED_RUNNER  := ./scripts/seed_runner

# Legacy seed script (kept for standalone / post-install use)
SEED_SCRIPT  := scripts/seed.sh

.DEFAULT_GOAL := help

# ── Phony targets ──────────────────────────────────────────────────────────────
.PHONY: build build-mcp build-tui install dist \
        db-init db-seed db-embeddings seed \
        clean test run tui \
        build-rust build-rust-static help

# ── Seed file list (sorted) ─────────────────────────────────────────────────
SEED_FILES := $(sort $(wildcard $(SEEDS_DIR)/*.sql))

# ── Build ──────────────────────────────────────────────────────────────────────

## build: Compile both binaries into ./dist/
build: build-mcp build-tui

## build-mcp: Compile the MCP server binary to ./dist/dev-forge-mcp
build-mcp:
	@mkdir -p $(DIST_DIR)
	$(GO_BUILD) -o $(DIST_DIR)/$(BINARY_MCP) ./cmd/dev-forge-mcp/
	@echo "Built $(DIST_DIR)/$(BINARY_MCP)"

## build-tui: Compile the CLI/TUI binary to ./dist/dev-forge
build-tui:
	@mkdir -p $(DIST_DIR)
	$(GO_BUILD) -o $(DIST_DIR)/$(BINARY_TUI) ./cmd/dev-forge/
	@echo "Built $(DIST_DIR)/$(BINARY_TUI)"

# ── Install ────────────────────────────────────────────────────────────────────

## install: Build and install both binaries to ~/.local/bin/
install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(DIST_DIR)/$(BINARY_MCP) $(INSTALL_DIR)/$(BINARY_MCP)
	cp $(DIST_DIR)/$(BINARY_TUI) $(INSTALL_DIR)/$(BINARY_TUI)
	chmod +x $(INSTALL_DIR)/$(BINARY_MCP) $(INSTALL_DIR)/$(BINARY_TUI)
	@if [ -f $(BIN_DIR)/devforge-imgproc ]; then \
		chmod +x $(BIN_DIR)/devforge-imgproc; \
		cp $(BIN_DIR)/devforge-imgproc $(INSTALL_DIR)/devforge-imgproc; \
		echo "Installed devforge-imgproc to $(INSTALL_DIR)/"; \
	else \
		echo "Warning: $(BIN_DIR)/devforge-imgproc not found — optimize_images / generate_favicon unavailable"; \
	fi
	@echo "Installed to $(INSTALL_DIR)/"
	@echo "Ensure $(INSTALL_DIR) is in your PATH: export PATH=\"\$$HOME/.local/bin:\$$PATH\""

# ── Distribution package ───────────────────────────────────────────────────────

## dist: Build binaries + fully initialize and seed the distribution DB
dist: build seed
	@if [ -f $(BIN_DIR)/devforge-imgproc ]; then \
		chmod +x $(BIN_DIR)/devforge-imgproc; \
		cp $(BIN_DIR)/devforge-imgproc $(DIST_DIR)/devforge-imgproc; \
	fi
	@echo "Distribution package ready in $(DIST_DIR)/"
	@echo "  binary : $(DIST_DIR)/$(BINARY_MCP)"
	@echo "  database: $(DB_PATH)"

# ── Database ───────────────────────────────────────────────────────────────────

## db-init: Create/migrate the libSQL DB using the project's own db.Open() / RunMigrations()
##          Writes to DB_PATH (default: dist/dev-forge.db). Idempotent.
db-init:
	@echo "Initializing database at $(DB_PATH)..."
	@mkdir -p $(DIST_DIR)
	$(GO_RUN) $(INIT_RUNNER) -db "$(DB_PATH)"

## db-seed: Apply all db/seeds/*.sql files to DB_PATH in numeric order (idempotent)
db-seed:
	@if [ ! -f "$(DB_PATH)" ]; then \
		echo "Database not found at $(DB_PATH) — run 'make db-init' first"; \
		exit 1; \
	fi
	@echo "Seeding database at $(DB_PATH)..."
	$(GO_RUN) $(SEED_RUNNER) -db "$(DB_PATH)" $(foreach f,$(SEED_FILES),-sql "$(f)")
	@echo "Seed complete."

## db-embeddings: Generate Ollama embeddings for rows with embedding IS NULL
db-embeddings:
	@bash $(SEED_SCRIPT) --embeddings-only \
		--db "$(DB_PATH)" \
		--ollama-url "$(OLLAMA_URL)" \
		--ollama-model "$(OLLAMA_MODEL)"

## seed: Full seed pipeline — db-init + db-seed + db-embeddings (idempotent)
seed: db-init db-seed db-embeddings

# ── Run ────────────────────────────────────────────────────────────────────────

## run: Build and run the MCP server (stdio transport)
run: build-mcp
	@echo "Starting MCP server (stdio transport)..."
	./$(DIST_DIR)/$(BINARY_MCP)

## tui: Build and run the CLI/TUI
tui: build-tui
	./$(DIST_DIR)/$(BINARY_TUI)

# ── Test ───────────────────────────────────────────────────────────────────────

## test: Run all tests with CGO_ENABLED=1 (required for libsql / FTS5)
test:
	$(GO_TEST) ./...

# ── Rust image-processing binary ───────────────────────────────────────────────

## build-rust: Build the imgproc Rust binary (dynamic, requires Rust toolchain)
build-rust:
	cd rust-imgproc && cargo build --release
	cp rust-imgproc/target/release/devforge-imgproc $(BIN_DIR)/devforge-imgproc
	chmod +x $(BIN_DIR)/devforge-imgproc
	@echo "Built $(BIN_DIR)/devforge-imgproc"

## build-rust-static: Build a fully static imgproc binary (musl, no system deps)
build-rust-static:
	cd rust-imgproc && cargo build --release --target x86_64-unknown-linux-musl
	cp rust-imgproc/target/x86_64-unknown-linux-musl/release/devforge-imgproc $(BIN_DIR)/devforge-imgproc
	chmod +x $(BIN_DIR)/devforge-imgproc
	@echo "Built static $(BIN_DIR)/devforge-imgproc"

# ── Clean ──────────────────────────────────────────────────────────────────────

## clean: Remove dist/ and compiled binaries
clean:
	rm -rf $(DIST_DIR)
	@echo "Cleaned $(DIST_DIR)/"

# ── Help ───────────────────────────────────────────────────────────────────────

## help: Show this help message
help:
	@echo "dev-forge-mcp — available make targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /' | column -t -s ':'
	@echo ""
	@echo "Variables (override on command line):"
	@echo "  OLLAMA_URL=$(OLLAMA_URL)"
	@echo "  OLLAMA_MODEL=$(OLLAMA_MODEL)"
	@echo "  DB_PATH=$(DB_PATH)"
	@echo "  INSTALL_DIR=$(INSTALL_DIR)"
