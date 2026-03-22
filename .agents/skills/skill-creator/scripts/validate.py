#!/usr/bin/env python3
"""
Validates a SKILL.md file against Claude Code skill best practices.
Usage: python validate.py <path-to-SKILL.md>
"""

import sys
import re
from pathlib import Path


RESERVED_WORDS = {"anthropic", "claude"}
MAX_NAME_LENGTH = 64
MAX_DESCRIPTION_LENGTH = 1024
MAX_BODY_LINES = 500
NAME_PATTERN = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
GERUND_HINT = re.compile(r"^(writing|processing|analyzing|generating|building|creating|managing|testing|reviewing|designing|running|checking|parsing|formatting|deploying|monitoring|extracting|converting|validating|documenting)-")
XML_TAG_PATTERN = re.compile(r"<[^>]+>")
WINDOWS_PATH_PATTERN = re.compile(r"[a-zA-Z]:\\|\\[a-zA-Z]")
DEEP_HEADING_PATTERN = re.compile(r"^#{4,}\s", re.MULTILINE)
TIME_SENSITIVE_PATTERN = re.compile(r"\b(antes de|después de|after|before)\b.{0,30}\b20\d{2}\b", re.IGNORECASE)


def parse_frontmatter(content: str) -> tuple[dict, str, list[str]]:
    errors = []
    frontmatter = {}
    body = content

    if not content.startswith("---"):
        errors.append("ERROR: El archivo no empieza con '---' (frontmatter YAML requerido).")
        return frontmatter, body, errors

    end = content.find("\n---", 3)
    if end == -1:
        errors.append("ERROR: Frontmatter no cerrado (falta '---' de cierre).")
        return frontmatter, body, errors

    raw_fm = content[3:end].strip()
    body = content[end + 4:].strip()

    for line in raw_fm.splitlines():
        if ":" in line:
            key, _, value = line.partition(":")
            frontmatter[key.strip()] = value.strip()

    return frontmatter, body, errors


def validate_name(name: str) -> list[str]:
    errors = []

    if not name:
        errors.append("ERROR [name]: Campo 'name' ausente o vacío.")
        return errors

    if len(name) > MAX_NAME_LENGTH:
        errors.append(f"ERROR [name]: '{name}' supera {MAX_NAME_LENGTH} caracteres ({len(name)}).")

    if not NAME_PATTERN.match(name):
        errors.append(f"ERROR [name]: '{name}' contiene caracteres inválidos. Solo minúsculas, números y guiones.")

    for word in RESERVED_WORDS:
        if word in name.split("-"):
            errors.append(f"ERROR [name]: '{name}' contiene la palabra reservada '{word}'.")

    if XML_TAG_PATTERN.search(name):
        errors.append(f"ERROR [name]: '{name}' contiene etiquetas XML.")

    if not GERUND_HINT.match(name):
        errors.append(f"WARN  [name]: '{name}' no usa forma de gerundio recomendada (writing-*, processing-*, analyzing-*, etc.).")

    return errors


def validate_description(description: str) -> list[str]:
    errors = []

    if not description:
        errors.append("ERROR [description]: Campo 'description' ausente o vacío.")
        return errors

    if len(description) > MAX_DESCRIPTION_LENGTH:
        errors.append(f"ERROR [description]: Supera {MAX_DESCRIPTION_LENGTH} caracteres ({len(description)}).")

    if XML_TAG_PATTERN.search(description):
        errors.append("ERROR [description]: Contiene etiquetas XML.")

    first_person = re.compile(r"\b(puedo|te ayudo|ayudarte|I can|I will)\b", re.IGNORECASE)
    if first_person.search(description):
        errors.append("WARN  [description]: Parece escrita en primera persona. Usar tercera persona ('Escribe...', 'Analiza...', 'Genera...').")

    trigger_words = ["úsalo", "usa este", "use this", "when", "cuando"]
    if not any(t in description.lower() for t in trigger_words):
        errors.append("WARN  [description]: No incluye indicación de cuándo usarlo ('Úsalo cuando...', 'Use when...').")

    return errors


def strip_code_blocks(text: str) -> str:
    """Remove fenced code block contents, keeping only the fence lines."""
    return re.sub(r"```[^\n]*\n.*?```", "```\n```", text, flags=re.DOTALL)


def validate_body(body: str) -> list[str]:
    errors = []
    lines = body.splitlines()

    if len(lines) > MAX_BODY_LINES:
        errors.append(f"WARN  [body]: {len(lines)} líneas supera el límite recomendado de {MAX_BODY_LINES}. Considera mover contenido a archivos referenciados.")

    body_no_code = strip_code_blocks(body)
    h1_matches = [l for l in body_no_code.splitlines() if re.match(r"^# [^#]", l)]
    if len(h1_matches) == 0:
        errors.append("WARN  [body]: No hay H1 (`# Título`). El cuerpo debería empezar con un H1.")
    elif len(h1_matches) > 1:
        errors.append(f"ERROR [body]: Hay {len(h1_matches)} encabezados H1 fuera de bloques de código. Solo se permite uno.")

    if DEEP_HEADING_PATTERN.search(body_no_code):
        errors.append("WARN  [body]: Contiene encabezados de nivel 4 o más (`####`). Evitar niveles profundos.")

    # Count unlabeled opening fences (not closing fences)
    opening_fences = re.findall(r"(?m)^```(\w*)$", body)
    unlabeled = [f for f in opening_fences if not f]
    # Subtract closing fences (every other unlabeled fence is a closing one)
    real_unlabeled = len(unlabeled) // 2
    if real_unlabeled > 0:
        errors.append(f"WARN  [body]: {real_unlabeled} bloque(s) de código sin lenguaje explícito. Agregar: ```typescript, ```bash, ```json, etc.")

    if WINDOWS_PATH_PATTERN.search(body_no_code):
        errors.append("WARN  [body]: Posibles rutas estilo Windows detectadas. Usar barras diagonales (`scripts/validate.py`).")

    if TIME_SENSITIVE_PATTERN.search(body_no_code):
        errors.append("WARN  [body]: Posible información sensible al tiempo detectada. Evitar referencias a fechas específicas.")

    bullet_styles = {"*": 0, "+": 0, "-": 0}
    for line in body_no_code.splitlines():
        m = re.match(r"^\s*([*+\-])\s", line)
        if m:
            bullet_styles[m.group(1)] += 1

    non_dash = bullet_styles["*"] + bullet_styles["+"]
    if non_dash > 0:
        errors.append(f"WARN  [body]: {non_dash} bullet(s) usan '*' o '+'. Usar '-' de forma consistente.")

    return errors


def main():
    if len(sys.argv) < 2:
        print("Uso: python validate.py <ruta-a-SKILL.md>")
        sys.exit(1)

    path = Path(sys.argv[1])
    if not path.exists():
        print(f"ERROR: Archivo no encontrado: {path}")
        sys.exit(1)

    content = path.read_text(encoding="utf-8")
    frontmatter, body, parse_errors = parse_frontmatter(content)

    all_errors = parse_errors[:]

    all_errors += validate_name(frontmatter.get("name", ""))
    all_errors += validate_description(frontmatter.get("description", ""))
    all_errors += validate_body(body)

    hard_errors = [e for e in all_errors if e.startswith("ERROR")]
    warnings = [e for e in all_errors if e.startswith("WARN")]

    print(f"\n{'='*60}")
    print(f"  Validando: {path}")
    print(f"{'='*60}\n")

    if hard_errors:
        print("ERRORES (deben corregirse):")
        for e in hard_errors:
            print(f"  {e}")
        print()

    if warnings:
        print("ADVERTENCIAS (revisar):")
        for w in warnings:
            print(f"  {w}")
        print()

    if not all_errors:
        print("OK — El skill pasa todas las validaciones.\n")
    elif not hard_errors:
        print(f"OK con advertencias — {len(warnings)} advertencia(s) a revisar.\n")
    else:
        print(f"FALLO — {len(hard_errors)} error(es) crítico(s) encontrado(s).\n")
        sys.exit(1)


if __name__ == "__main__":
    main()
