---
name: skill-creator
description: Diseña, redacta y valida archivos SKILL.md para Claude Code siguiendo las buenas prácticas oficiales. Úsalo cuando el usuario quiera crear un nuevo skill, mejorar uno existente, revisar su estructura o aprender cómo construir skills efectivos para Claude Code.
---

# Skill Creator

Skill para crear Skills de Claude Code bien estructurados, concisos y efectivos.

---

## Principios fundamentales

**Concisión sobre exhaustividad**: Claude ya es inteligente. Solo incluye contexto que no tenga.

**Descripción como trigger**: La `description` es lo que activa el skill. Debe responder "¿qué hace?" y "¿cuándo usarlo?" con términos clave específicos — en tercera persona.

**Divulgación progresiva**: SKILL.md como índice. Detalles en archivos referenciados, cargados solo cuando se necesitan. Sin penalización de contexto para archivos no leídos.

**Límite de 500 líneas**: Si el cuerpo de SKILL.md supera ese límite, mover contenido a archivos referenciados.

---

## Flujo de trabajo

### 1 — Recopilar información

Pregunta solo lo que no esté claro:

- ¿Qué tarea concreta resuelve el skill?
- ¿Qué palabras o contexto disparan su uso?
- ¿Necesita referencias, plantillas o scripts de apoyo?

### 2 — Definir nombre y descripción

- **Nombre**: gerundio en kebab-case, solo minúsculas/números/guiones, máximo 64 caracteres.
  - Bueno: `writing-skills`, `processing-pdfs`, `analyzing-logs`
  - Evitar: `helper`, `utils`, `tools`, palabras reservadas (`anthropic`, `claude`)
- **Descripción**: tercera persona, específica, máximo 1024 caracteres. Incluye qué hace Y cuándo usarlo.

Consulta [reference/anatomy.md](reference/anatomy.md) para las restricciones completas del frontmatter.

### 3 — Elegir arquitectura

| Caso | Estructura |
|------|------------|
| Skill simple (< 500 líneas) | Solo `SKILL.md` |
| Con docs adicionales | `SKILL.md` + `reference/*.md` |
| Con plantillas de output | `SKILL.md` + `templates/*.md` |
| Con scripts ejecutables | `SKILL.md` + `scripts/*.py` |
| Combinado | Los cuatro subdirectorios |

Regla: referencias a **máximo 1 nivel de profundidad** desde SKILL.md.

### 4 — Redactar el cuerpo

Usa la plantilla de inicio rápido en [templates/skill-template.md](templates/skill-template.md).

Secciones recomendadas según el tipo de skill:

- **Flujo de trabajo**: pasos numerados con puntos de decisión.
- **Reglas / guidelines**: instrucciones específicas que Claude debe seguir.
- **Checklist de validación**: Claude la aplica antes de entregar el output.
- **Ejemplos**: pares input/output concretos cuando el estilo del output importa.

### 5 — Escribir el archivo

**Usa la herramienta Write** para crear el archivo en `.agents/skills/<nombre-skill>/SKILL.md`.

Reglas críticas de entrega:

- **Nunca** muestres el contenido del SKILL.md dentro de un bloque de código (triple o cuádruple backtick).
- El contenido del archivo se escribe directamente con Write — no como texto de respuesta.
- Si necesitas crear subdirectorios (`reference/`, `templates/`, `scripts/`), crea también esos archivos con Write.

Después de crear todos los archivos del skill, ejecuta el script de enlace:

```bash
./scripts/link-skills.sh
```

Esto crea automáticamente el symlink `.claude/skills/<nombre-skill> → ../../.agents/skills/<nombre-skill>` para que Claude Code lo descubra.

### 6 — Validar

Ejecuta el script de validación automática:

```bash
python3 .agents/skills/writing-skills/scripts/validate.py .agents/skills/<nombre-skill>/SKILL.md
```

Luego aplica el checklist manual en [reference/checklist.md](reference/checklist.md).

### 7 — Registrar en CLAUDE.md

Después de crear o renombrar un skill, añade (o actualiza) su entrada en la tabla **Skills** de `CLAUDE.md`:

```markdown
| `nombre-skill` | Cuándo usarlo (una línea) |
```

Usa la herramienta Edit para insertar la fila en orden alfabético dentro de la tabla existente. Si el skill ya existe, actualiza la descripción de la columna "When to use".

---

## Cuándo agregar scripts

Los scripts son útiles cuando la operación es:

- **Frágil o propensa a errores**: validación de formato, parsing de YAML.
- **Determinista**: siempre produce el mismo resultado para el mismo input.
- **Repetitiva**: se ejecutará en múltiples archivos o sesiones.

Instrucción en SKILL.md: "Ejecuta `scripts/validate.py`" (no "lee `scripts/validate.py`").

---

## Cuándo agregar referencias a otros skills

Un skill puede mencionar otros skills del mismo proyecto:

```text
Antes de redactar el cuerpo, aplica las reglas de
[writing-markdown](../writing-markdown/SKILL.md) para la sintaxis.
```

Úsalo cuando el skill se apoya en convenciones ya definidas en otro skill — evita duplicar instrucciones.

---

## Skills de referencia en este proyecto

| Skill | Descripción |
|-------|-------------|
| [write-prp](../write-prp/SKILL.md) | Estructura G3 (Guidelines · Guidance · Guardrails) con workflow en capas |
| [writing-markdown](../writing-markdown/SKILL.md) | Skill simple con checklist de autovalidación |
---

## Checklist rápido

- [ ] `name` en gerundio, kebab-case, sin palabras reservadas.
- [ ] `description` en tercera persona con qué hace + cuándo usarlo.
- [ ] Cuerpo bajo 500 líneas.
- [ ] Sin saltos de nivel en encabezados.
- [ ] Referencias a máximo 1 nivel de profundidad.
- [ ] Archivo creado con Write en `.agents/skills/<nombre-skill>/` (no mostrado como bloque de código).
- [ ] `./scripts/link-skills.sh` ejecutado — symlink en `.claude/skills/` creado.
- [ ] Script de validación ejecutado sin errores.
- [ ] Checklist completo en [reference/checklist.md](reference/checklist.md) aplicado.
- [ ] Entrada agregada o actualizada en la tabla Skills de `CLAUDE.md`.
