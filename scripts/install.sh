#!/usr/bin/env bash
# install.sh — Install devforge to ~/.local/share/devforge/versions/X.Y.Z/
# and create symlinks in ~/.local/bin/
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# ── Version ────────────────────────────────────────────────────────────────────
VERSION_FILE="${PROJECT_DIR}/VERSION"
if [ -f "$VERSION_FILE" ]; then
    VERSION="$(tr -d '[:space:]' < "$VERSION_FILE")"
else
    VERSION="0.0.0"
fi

SHARE_BASE="${HOME}/.local/share/devforge"
SHARE_DIR="${SHARE_BASE}/versions/${VERSION}"
LINK_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/devforge"
CONFIG_FILE="${CONFIG_DIR}/config.json"
DIST_DIR="${PROJECT_DIR}/dist"

RED='\033[0;31m'
YLW='\033[1;33m'
GRN='\033[0;32m'
CYN='\033[0;36m'
BLD='\033[1m'
RST='\033[0m'

echo ""
echo -e "${BLD}DevForge — Installer v${VERSION}${RST}"
echo -e "${CYN}──────────────────────────────────────────${RST}"
echo ""

# ── Build ──────────────────────────────────────────────────────────────────────
echo -e "${BLD}Step 1/5 — Building binaries...${RST}"
cd "$PROJECT_DIR"
CGO_ENABLED=1 go build -o "${DIST_DIR}/devforge-mcp" ./cmd/devforge-mcp/
CGO_ENABLED=1 go build -o "${DIST_DIR}/devforge"     ./cmd/devforge/
echo -e "  ${GRN}✓${RST} devforge-mcp"
echo -e "  ${GRN}✓${RST} devforge"

# ── Create share dir and copy files ───────────────────────────────────────────
echo ""
echo -e "${BLD}Step 2/5 — Installing to ${SHARE_DIR}...${RST}"
mkdir -p "${SHARE_DIR}"

rm -f "${SHARE_DIR}/devforge-mcp" "${SHARE_DIR}/devforge"
cp "${DIST_DIR}/devforge-mcp" "${SHARE_DIR}/devforge-mcp"
cp "${DIST_DIR}/devforge"     "${SHARE_DIR}/devforge"
chmod +x "${SHARE_DIR}/devforge-mcp" "${SHARE_DIR}/devforge"
echo -e "  ${GRN}✓${RST} devforge-mcp"
echo -e "  ${GRN}✓${RST} devforge"

if [ -f "${DIST_DIR}/dpf" ]; then
    rm -f "${SHARE_DIR}/dpf"
    cp "${DIST_DIR}/dpf" "${SHARE_DIR}/dpf"
    chmod +x "${SHARE_DIR}/dpf"
    echo -e "  ${GRN}✓${RST} dpf"
elif [ -f "${PROJECT_DIR}/bin/dpf" ]; then
    cp "${PROJECT_DIR}/bin/dpf" "${SHARE_DIR}/dpf"
    chmod +x "${SHARE_DIR}/dpf"
    echo -e "  ${GRN}✓${RST} dpf (from bin/)"
else
    echo -e "  ${YLW}⚠${RST}  dpf not found — media tools unavailable"
    echo -e "  ${YLW}→${RST}  Run: bash scripts/install-dpf.sh"
fi

# Copy or initialize database
DB_TARGET="${SHARE_DIR}/devforge.db"
if [ -f "${DIST_DIR}/devforge.db" ]; then
    cp "${DIST_DIR}/devforge.db" "${DB_TARGET}"
    echo -e "  ${GRN}✓${RST} devforge.db (from dist/)"
else
    echo -e "  ${YLW}⚠${RST}  dist/devforge.db not found — initializing empty database"
    CGO_ENABLED=1 go run "${PROJECT_DIR}/scripts/init_db_runner" -db "${DB_TARGET}"
    echo -e "  ${GRN}✓${RST} devforge.db (initialized)"
fi

# Update 'current' symlink
ln -sfn "versions/${VERSION}" "${SHARE_BASE}/current"
echo -e "  ${GRN}✓${RST} ${SHARE_BASE}/current -> versions/${VERSION}"

# ── Optional seed ─────────────────────────────────────────────────────────────
echo ""
echo -e "${BLD}Step 3/5 — Database seed${RST}"
echo -e "  Database: ${BLD}${DB_TARGET}${RST}"
echo ""
echo -n "Apply seed data (patterns, architectures, palettes)? [y/N] "
read -r SEED_ANSWER
if [[ "${SEED_ANSWER}" =~ ^[Yy]$ ]]; then
    SEEDS_DIR="${PROJECT_DIR}/db/seeds"
    SEED_FILES=$(ls "${SEEDS_DIR}"/*.sql 2>/dev/null | sort)
    if [ -z "${SEED_FILES}" ]; then
        echo -e "  ${YLW}⚠${RST}  No seed files found in ${SEEDS_DIR}"
    else
        CGO_ENABLED=1 go run "${PROJECT_DIR}/scripts/seed_runner" \
            -db "${DB_TARGET}" \
            $(echo "${SEED_FILES}" | xargs -I{} echo -sql "{}")
        echo -e "  ${GRN}✓${RST} Seeds applied"
    fi
else
    echo -e "  ${YLW}skipped${RST}"
fi

# ── Symlinks in ~/.local/bin ───────────────────────────────────────────────────
echo ""
echo -e "${BLD}Step 4/5 — Creating symlinks in ${LINK_DIR}...${RST}"
mkdir -p "${LINK_DIR}"

for BIN in devforge-mcp devforge; do
    TARGET="${SHARE_DIR}/${BIN}"
    LINK="${LINK_DIR}/${BIN}"
    if [ -f "${TARGET}" ]; then
        ln -sf "${TARGET}" "${LINK}"
        echo -e "  ${GRN}✓${RST} ${LINK} -> ${TARGET}"
    fi
done

if [ -f "${SHARE_DIR}/dpf" ]; then
    ln -sf "${SHARE_DIR}/dpf" "${LINK_DIR}/dpf"
    echo -e "  ${GRN}✓${RST} ${LINK_DIR}/dpf -> ${SHARE_DIR}/dpf"
fi

# ── Initial config ─────────────────────────────────────────────────────────────
echo ""
echo -e "${BLD}Step 5/5 — Configuration${RST}"
mkdir -p "${CONFIG_DIR}"
if [ ! -f "${CONFIG_FILE}" ]; then
    cat > "${CONFIG_FILE}" <<'EOF'
{
  "gemini_api_key": "",
  "ollama_url": "http://localhost:11434",
  "embedding_model": "nomic-embed-text"
}
EOF
    chmod 600 "${CONFIG_FILE}"
    echo -e "  ${GRN}✓${RST} Created ${CONFIG_FILE}"
else
    echo -e "  ${YLW}kept${RST}   ${CONFIG_FILE} (already exists)"
fi

# ── Summary ────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GRN}${BLD}DevForge v${VERSION} installed successfully.${RST}"
echo ""
echo -e "${BLD}Locations:${RST}"
echo "  Binaries : ${SHARE_DIR}/"
echo "  Database : ${DB_TARGET}"
echo "  Config   : ${CONFIG_FILE}"
echo "  Symlinks : ${LINK_DIR}/devforge-mcp, devforge"
echo ""
echo "Ensure ${LINK_DIR} is in your PATH:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
