#!/usr/bin/env bash
# bump-version.sh — Atomically update all version strings across the project.
#
# Usage:
#   bash scripts/bump-version.sh X.Y.Z
#
# Files updated:
#   VERSION                           plain-text version
#   internal/version/version.go       const Current
#   cmd/devforge-mcp/main.go          NewMCPServer second argument
#   README.md                         version badge
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

NEW_VERSION="${1:-}"

if [ -z "$NEW_VERSION" ]; then
  echo "Usage: bash scripts/bump-version.sh X.Y.Z" >&2
  exit 1
fi

if ! printf '%s' "$NEW_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "Version must follow semver X.Y.Z. Got: $NEW_VERSION" >&2
  exit 1
fi

OLD_VERSION="$(tr -d '[:space:]' < "${PROJECT_DIR}/VERSION")"

if [ "$OLD_VERSION" = "$NEW_VERSION" ]; then
  echo "Already at version $NEW_VERSION — nothing to do."
  exit 0
fi

# Portable in-place sed (Linux vs macOS)
if [[ "$(uname -s)" == "Darwin" ]]; then
  SED() { sed -i '' "$@"; }
else
  SED() { sed -i "$@"; }
fi

echo "Bumping $OLD_VERSION → $NEW_VERSION"

# 1. VERSION file
printf '%s\n' "$NEW_VERSION" > "${PROJECT_DIR}/VERSION"

# 2. internal/version/version.go — const Current
VERSION_GO="${PROJECT_DIR}/internal/version/version.go"
SED "s/const Current = \"${OLD_VERSION}\"/const Current = \"${NEW_VERSION}\"/" "$VERSION_GO"

if ! grep -q "\"${NEW_VERSION}\"" "$VERSION_GO"; then
  echo "ERROR: failed to update $VERSION_GO — const Current not found for ${OLD_VERSION}" >&2
  exit 1
fi

# 3. cmd/devforge-mcp/main.go — NewMCPServer second argument
MCP_MAIN="${PROJECT_DIR}/cmd/devforge-mcp/main.go"
SED "s/NewMCPServer(\"devforge\", \"${OLD_VERSION}\"/NewMCPServer(\"devforge\", \"${NEW_VERSION}\"/" "$MCP_MAIN"

if ! grep -q "\"${NEW_VERSION}\"" "$MCP_MAIN"; then
  echo "ERROR: failed to update $MCP_MAIN — NewMCPServer version not found for ${OLD_VERSION}" >&2
  exit 1
fi

# 4. README.md — version badge
README="${PROJECT_DIR}/README.md"
SED "s/version-${OLD_VERSION}-blue/version-${NEW_VERSION}-blue/" "$README"

echo ""
echo "Updated:"
echo "  VERSION"
echo "  internal/version/version.go"
echo "  cmd/devforge-mcp/main.go"
echo "  README.md"
echo ""
echo "Next steps:"
echo "  go test ./..."
echo "  git add VERSION internal/version/version.go cmd/devforge-mcp/main.go README.md"
echo "  git commit -m \"chore: bump to v${NEW_VERSION}\""
echo "  git tag v${NEW_VERSION}"
echo "  git push origin main && git push origin v${NEW_VERSION}"
