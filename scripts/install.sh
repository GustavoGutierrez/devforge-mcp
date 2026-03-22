#!/usr/bin/env bash
# install.sh — Build and install dev-forge binaries to ~/.local/bin
set -euo pipefail

INSTALL_DIR="${HOME}/.local/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Building dev-forge-mcp and dev-forge..."
cd "$PROJECT_DIR"

CGO_ENABLED=1 go build -o bin/dev-forge-mcp ./cmd/dev-forge-mcp/
CGO_ENABLED=1 go build -o bin/dev-forge     ./cmd/dev-forge/

mkdir -p "$INSTALL_DIR"

cp bin/dev-forge-mcp  "$INSTALL_DIR/dev-forge-mcp"
cp bin/dev-forge      "$INSTALL_DIR/dev-forge"

if [ -f bin/devforge-imgproc ]; then
    chmod +x bin/devforge-imgproc
    cp bin/devforge-imgproc "$INSTALL_DIR/devforge-imgproc"
    echo "Installed devforge-imgproc to $INSTALL_DIR/"
else
    echo "Warning: bin/devforge-imgproc not found — optimize_images and generate_favicon will be unavailable"
fi

chmod +x "$INSTALL_DIR/dev-forge-mcp" "$INSTALL_DIR/dev-forge"

echo "Installed:"
echo "  $INSTALL_DIR/dev-forge-mcp"
echo "  $INSTALL_DIR/dev-forge"
echo ""
echo "Ensure $INSTALL_DIR is in your PATH:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "Initial config setup:"
echo "  mkdir -p ~/.config/dev-forge"
echo "  cat > ~/.config/dev-forge/config.json <<'EOF'"
echo '  {"gemini_api_key":"","ollama_url":"http://localhost:11434","embedding_model":"nomic-embed-text"}'
echo "  EOF"
echo "  chmod 600 ~/.config/dev-forge/config.json"
