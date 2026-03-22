#!/usr/bin/env bash
# link-skills.sh — Crea symlinks en .claude/skills/ apuntando a .agents/skills/
# Idempotente: omite links que ya existen y correctos, reemplaza los rotos.
#
# Uso:
#   ./scripts/link-skills.sh
#
# Resultado:
#   .claude/skills/<skill> -> ../../.agents/skills/<skill>
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

AGENTS_DIR="$REPO_ROOT/.agents/skills"
CLAUDE_DIR="$REPO_ROOT/.claude/skills"

if [[ ! -d "$AGENTS_DIR" ]]; then
    echo "Error: directorio fuente no encontrado: $AGENTS_DIR" >&2
    exit 1
fi

mkdir -p "$CLAUDE_DIR"

linked=0
skipped=0
replaced=0

for skill_path in "$AGENTS_DIR"/*/; do
    [[ -d "$skill_path" ]] || continue
    skill="$(basename "$skill_path")"
    target="../../.agents/skills/$skill"
    link="$CLAUDE_DIR/$skill"

    if [[ -L "$link" ]]; then
        current="$(readlink "$link")"
        if [[ "$current" == "$target" ]]; then
            echo "  ok  $skill"
            ((skipped++)) || true
            continue
        else
            echo "  fix $skill  (era: $current)"
            rm "$link"
        fi
    elif [[ -e "$link" ]]; then
        echo "  !!  $skill es un archivo/directorio real — se omite para no destruir datos" >&2
        continue
    fi

    ln -s "$target" "$link"
    echo "  +   $skill"
    ((linked++)) || true
done

echo ""
echo "Listo: $linked creados, $replaced reemplazados, $skipped ya correctos."
echo "Symlinks en: $CLAUDE_DIR"
