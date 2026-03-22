# Anatomía de un SKILL.md

## Tabla de contenidos

- Frontmatter YAML
- Estructura del cuerpo
- Patrones de divulgación progresiva
- Límites y restricciones

---

## Frontmatter YAML (obligatorio)

```yaml
---
name: nombre-en-gerundio
description: Qué hace este skill y cuándo usarlo — incluye términos clave específicos.
---
```

### Restricciones de `name`

| Regla                             | Ejemplo válido     | Ejemplo inválido                  |
| --------------------------------- | ------------------ | --------------------------------- |
| Solo minúsculas, números, guiones | `writing-skills`   | `Writing_Skills`                  |
| Máximo 64 caracteres              | `processing-pdfs`  | _(string de 65+ chars)_           |
| Sin etiquetas XML                 | `analyzing-logs`   | `<skill>analyzing</skill>`        |
| Sin palabras reservadas           | `writing-docs`     | `claude-helper`, `anthropic-tool` |
| Forma recomendada: gerundio       | `writing-markdown` | `markdown`, `helper`, `utils`     |

### Restricciones de `description`

| Regla                  | Correcto                                        | Incorrecto                     |
| ---------------------- | ----------------------------------------------- | ------------------------------ |
| Tercera persona        | "Escribe archivos .md..."                       | "Puedo ayudarte a escribir..." |
| Máximo 1024 caracteres | _(cualquier texto bajo el límite)_              | _(texto de 1025+ chars)_       |
| No vacío               | "Diseña y valida SKILL.md..."                   | `""`                           |
| Sin etiquetas XML      | texto plano                                     | `<b>Escribe</b> archivos`      |
| Incluye cuándo usarlo  | "...Úsalo cuando el usuario pida crear skills." | Solo describe sin trigger      |

---

## Estructura del cuerpo

### Título H1 (obligatorio, único)

```markdown
# Nombre del Skill
```

Solo un H1 en todo el archivo. No usar H1 en archivos de referencia.

### Secciones con H2

```markdown
## Principios

## Flujo de trabajo

## Reglas

## Checklist de validación
```

### Sub-secciones con H3

```markdown
### Paso 1 — Nombre del paso
```

No usar `####` ni niveles más profundos.

### Orden recomendado del cuerpo

1. Descripción de una línea (qué resuelve).
2. Principios o contexto clave (solo lo que Claude no sabe).
3. Flujo de trabajo (pasos numerados).
4. Reglas específicas del dominio.
5. Ejemplos concretos (si el estilo del output importa).
6. Checklist de validación.
7. Referencias a archivos adicionales.

---

## Patrones de divulgación progresiva

Solo los metadatos (`name` + `description`) se precargan. Claude lee SKILL.md cuando el skill se activa, y lee archivos adicionales solo si los necesita.

### Patrón 1 — Guía con referencias

```markdown
## Uso avanzado

Para casos de uso complejos, consulta [reference/advanced.md](reference/advanced.md).
Para la plantilla de output, usa [templates/output.md](templates/output.md).
```

### Patrón 2 — Organización por dominio

```bash
skills/bigquery-analysis/
├── SKILL.md                  # overview + navegación
└── reference/
    ├── finance.md            # métricas financieras
    ├── sales.md              # pipeline y oportunidades
    └── product.md            # uso de API
```

Claude solo lee `finance.md` si la pregunta es sobre finanzas.

### Patrón 3 — Scripts ejecutables

````markdown
## Validación

Ejecuta el validador:

```bash
python scripts/validate.py input.md
```
````

Claude ejecuta el script sin cargar su contenido en contexto.

### Regla de profundidad

```bash

✅ SKILL.md → reference/guide.md (1 nivel)
❌ SKILL.md → reference/guide.md → reference/details.md (2 niveles — evitar)

```

Todos los archivos referenciados deben apuntar directamente desde SKILL.md.

---

## Límites y restricciones

| Elemento | Límite | Acción si se supera |
|----------|--------|---------------------|
| Cuerpo de SKILL.md | 500 líneas | Mover contenido a archivos referenciados |
| Archivos de referencia > 100 líneas | — | Agregar tabla de contenidos al inicio |
| Profundidad de referencias | 1 nivel desde SKILL.md | No anidar referencias |
| Scripts | Sin límite de tamaño | Claude los ejecuta, no los lee en contexto |
