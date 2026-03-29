#!/usr/bin/env bash
# install-dpf.sh — Download the pre-built DevPixelForge binary
# Usage:
#   bash scripts/install-dpf.sh          # downloads latest release
#   bash scripts/install-dpf.sh v0.2.0  # downloads specific version
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="${PROJECT_DIR}/bin"
REPO="GustavoGutierrez/devpixelforge"

if [[ -t 1 ]]; then
    GRN='\033[0;32m'; YLW='\033[1;33m'; RED='\033[0;31m'; RST='\033[0m'
else
    GRN=''; YLW=''; RED=''; RST=''
fi

info()    { echo -e "${GRN}→${RST} $*"; }
warn()    { echo -e "${YLW}!${RST} $*"; }
error()   { echo -e "${RED}✗${RST} $*" >&2; }

fetch_latest() {
    local tag
    tag=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"v\?\([^"]*\)".*/\1/')
    if [ -z "$tag" ]; then
        error "Could not fetch latest release tag from GitHub API."
        exit 1
    fi
    echo "$tag"
}

download_dpf() {
    local version="$1"
    local url="https://github.com/${REPO}/releases/download/${version}/dpf"
    local dest="${BIN_DIR}/dpf"

    mkdir -p "${BIN_DIR}"

    info "Downloading DevPixelForge v${version}..."
    info "From: ${url}"

    if command -v curl &>/dev/null; then
        curl -fsSL -o "${dest}" "$url"
    elif command -v wget &>/dev/null; then
        wget -q -O "${dest}" "$url"
    else
        error "Neither curl nor wget found — cannot download."
        exit 1
    fi

    chmod +x "${dest}"
    info "Installed: ${dest}"
}

if [ $# -gt 0 ]; then
    VERSION="$1"
else
    info "Fetching latest release tag..."
    VERSION=$(fetch_latest)
    info "Latest version: ${VERSION}"
fi

# Strip leading 'v' if present
VERSION="${VERSION#v}"

download_dpf "${VERSION}"

echo ""
info "Done. Run 'chmod +x bin/dpf' to ensure it is executable."
info "See docs/install.md for the full installation guide."
