# Checklist de validación de SKILL.md

Aplica este checklist antes de entregar o publicar cualquier skill.

---

## Frontmatter

- [ ] Tiene bloque `---` de apertura y cierre.
- [ ] Campo `name` presente y en kebab-case (solo minúsculas, números, guiones).
- [ ] `name` no supera 64 caracteres.
- [ ] `name` no contiene palabras reservadas (`anthropic`, `claude`).
- [ ] `name` en forma de gerundio (`writing-*`, `processing-*`, `analyzing-*`).
- [ ] Campo `description` presente y no vacío.
- [ ] `description` no supera 1024 caracteres.
- [ ] `description` escrita en tercera persona.
- [ ] `description` incluye qué hace el skill Y cuándo usarlo.
- [ ] `description` incluye términos clave específicos para activación.
- [ ] Ningún campo contiene etiquetas XML.

---

## Cuerpo del documento

- [ ] Exactamente un H1 (`#`) al inicio del cuerpo.
- [ ] Encabezados no saltan niveles (`##` → `###`, nunca `##` → `####`).
- [ ] Cuerpo bajo 500 líneas.
- [ ] Sin información sensible al tiempo ("antes de agosto de 2025...").
- [ ] Terminología consistente en todo el archivo.
- [ ] Listas usan `-` para bullets (no `*` ni `+`).
- [ ] Bloques de código con lenguaje explícito (` ```typescript `, ` ```bash `, etc.).
- [ ] Sin rutas estilo Windows (`scripts\validate.py` → usar `scripts/validate.py`).

---

## Estructura y arquitectura

- [ ] Referencias a archivos adicionales a máximo 1 nivel de profundidad desde SKILL.md.
- [ ] Archivos de referencia > 100 líneas tienen tabla de contenidos al inicio.
- [ ] Scripts mencionados con instrucción "Ejecuta" (no "Lee").
- [ ] Si hay templates: SKILL.md los referencia con ruta relativa.
- [ ] Si hay referencias a otros skills: rutas relativas correctas (`../otro-skill/SKILL.md`).

---

## Calidad del contenido

- [ ] No incluye explicaciones que Claude ya sabe (qué es un JSON, qué es un endpoint REST, etc.).
- [ ] Flujo de trabajo con pasos numerados claros.
- [ ] Ejemplos concretos si el estilo del output importa.
- [ ] Checklist de autovalidación incluido para que Claude se autocorrija.
- [ ] Skill probado con al menos un caso de uso real antes de publicar.

---

## Validación automática

```bash
python .claude/skills/writing-skills/scripts/validate.py <ruta-al-SKILL.md>
```

El script verifica frontmatter, longitud del cuerpo y reglas de sintaxis básicas.
