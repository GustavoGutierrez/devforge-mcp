---
name: publish-release
description: >
  Step-by-step workflow for publishing a new DevForge version to GitHub Releases and Homebrew tap.
  Trigger: When the user asks to release a new version, bump the version, publish to GitHub,
  push a new release, create a release, "subir versión", "nueva versión", "nuevo release",
  "publicar versión", or "crear release".
license: Apache-2.0
metadata:
  author: GustavoGutierrez
  version: "1.0"
---

## When to Use

Load this skill when:
- User asks to release a new version, bump version, or publish to GitHub/Homebrew
- User says "subir versión", "nueva versión", "nuevo release", "publicar versión", or "crear release"
- User asks to tag a release or increment the version number

---

## Critical Patterns

- **`gh release create` uses `--notes`, NOT `--body`** — `--body` is an unknown flag and will crash the CI.
- **`VERSION` is the single source of truth** — always read it first; never hardcode the version.
- **`v` prefix required in workflow dispatch** — pass `version=vX.Y.Z`, not `X.Y.Z`.
- **Use `Edit` tool for Go/Ruby files** — do NOT use bash `sed` for `main.go` or `devforge.rb`.
- **PR branch must differ from base** — CI pushes to `update-vX.Y.Z`, PR targets `homebrew-tap`. Never push directly to `homebrew-tap`.
- **`CGO_ENABLED=1` required** — `go build ./...` without it will fail due to `go-sqlite3`.
- **Release notes must be English Markdown** — no plain text, no other language.
- **4 files must be updated every release**: `VERSION`, `README.md`, `cmd/devforge-mcp/main.go`, `Formula/devforge.rb`.

---

## Semantic Versioning Rules

| Change type | Segment | Example |
|-------------|---------|---------|
| Breaking change to MCP API, protocol, or config schema | **MAJOR** | `1.x.x → 2.0.0` |
| New MCP tool, CLI command, dpf job type (backwards-compatible) | **MINOR** | `1.0.x → 1.1.0` |
| Bug fix, refactor, docs update, dependency update, CI fix | **PATCH** | `1.0.1 → 1.0.2` |

---

## Files Updated Every Release

| File | What to change |
|------|---------------|
| `VERSION` | Plain semver string (no `v` prefix): `1.0.2` |
| `README.md` | Badge URL: `version-X.Y.Z-blue.svg` |
| `cmd/devforge-mcp/main.go` | `mcpserver.NewMCPServer("devforge", "X.Y.Z", ...)` |
| `Formula/devforge.rb` | `version "X.Y.Z"` |

---

## Step-by-Step Release Process

### Step 1 — Determine the new version

```bash
cat VERSION
```

Apply SemVer rules. Decide the new version (e.g., `1.0.2`).

---

### Step 2 — Update all version references

Use the `Edit` tool for each file:

- `VERSION` → plain semver string, no `v` prefix
- `README.md` → badge `version-OLD-blue.svg` → `version-NEW-blue.svg`
- `cmd/devforge-mcp/main.go` → version string in `NewMCPServer`
- `Formula/devforge.rb` → `version "OLD"` → `version "NEW"`

---

### Step 3 — Verify the build compiles clean

```bash
CGO_ENABLED=1 go build ./...
```

Fix any errors before continuing.

---

### Step 4 — Write the release notes

Use this template (English Markdown only, include only relevant sections):

```markdown
## What's Changed

### Breaking Changes
- Description and migration path (MAJOR only)

### New Features
- Description of new tool/command (MINOR only)

### Bug Fixes
- Fix description (#issue or short description)

### Improvements
- Description of improvement

### Refactors
- Description of refactor

### Dependency Updates
- Updated X from vA to vB

---

**Full Changelog**: https://github.com/GustavoGutierrez/devforge-mcp/compare/vPREV...vNEW
```

---

### Step 5 — Commit and push to main

```bash
git add VERSION README.md cmd/devforge-mcp/main.go Formula/devforge.rb
git commit -m "chore(release): bump version to vX.Y.Z

<one-line summary of what changed>"
git push origin main
```

---

### Step 6 — Trigger the CI workflow

```bash
gh workflow run homebrew.yml \
  -f version=vX.Y.Z \
  --repo GustavoGutierrez/devforge-mcp
```

---

### Step 7 — Monitor the workflow

```bash
sleep 5 && gh run list --repo GustavoGutierrez/devforge-mcp --limit 3
gh run watch <RUN_ID> --repo GustavoGutierrez/devforge-mcp
```

| Job | What it does |
|-----|-------------|
| `prepare-release` | Creates GitHub Release at `vX.Y.Z` |
| `build-linux` | Builds binaries, packages `.tar.gz`, uploads to release |
| `update-formula` | Rewrites `Formula/devforge.rb`, pushes `update-vX.Y.Z`, opens PR |

---

### Step 8 — Verify the release and PR

```bash
gh release view vX.Y.Z --repo GustavoGutierrez/devforge-mcp
gh pr list --repo GustavoGutierrez/devforge-mcp --state open
```

Expected: release has `devforge-X.Y.Z.linux-amd64.tar.gz`; PR targets `homebrew-tap`.

---

## Commands

```bash
# Read current version
cat VERSION

# Build check (required before release)
CGO_ENABLED=1 go build ./...

# Trigger release CI
gh workflow run homebrew.yml -f version=vX.Y.Z --repo GustavoGutierrez/devforge-mcp

# Monitor runs
gh run list --repo GustavoGutierrez/devforge-mcp --limit 3
gh run watch <RUN_ID> --repo GustavoGutierrez/devforge-mcp

# Verify release
gh release view vX.Y.Z --repo GustavoGutierrez/devforge-mcp

# Verify formula PR
gh pr list --repo GustavoGutierrez/devforge-mcp --state open
```

---

## Complete Checklist

```
[ ] Determined correct SemVer segment (MAJOR / MINOR / PATCH)
[ ] VERSION updated (no v prefix)
[ ] README.md badge updated
[ ] cmd/devforge-mcp/main.go version string updated
[ ] Formula/devforge.rb version updated
[ ] CGO_ENABLED=1 go build ./... passes clean
[ ] Release notes written in English Markdown
[ ] Committed: chore(release): bump version to vX.Y.Z
[ ] Pushed to main
[ ] Triggered: gh workflow run homebrew.yml -f version=vX.Y.Z
[ ] All 3 CI jobs passed (prepare-release, build-linux, update-formula)
[ ] GitHub Release vX.Y.Z exists with linux-amd64.tar.gz asset
[ ] Homebrew formula PR opened targeting homebrew-tap branch
```

---

## Resources

- **CI pipeline**: See [.github/workflows/homebrew.yml](../../../.github/workflows/homebrew.yml)
- **Homebrew formula**: See [Formula/devforge.rb](../../../Formula/devforge.rb)
- **Release conventions**: See [AGENTS.md](../../../AGENTS.md) — Homebrew Tap section
