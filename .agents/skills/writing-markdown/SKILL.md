---
name: writing-markdown
description: Escribe, redacta o corrige archivos .md (PRPs, READMEs, specs, docs) aplicando buenas prácticas de sintaxis Markdown y estilo de escritura técnica. Úsalo cuando el usuario pida crear, mejorar o revisar cualquier archivo Markdown.
---

# Writing Markdown

Skill para generar y revisar archivos `.md` con sintaxis limpia, estructura consistente y estilo de escritura técnica claro.

---

## Reglas de sintaxis

### Encabezados

- Usa `#` solo para el título del archivo — nunca en el cuerpo del documento.
- `##` para secciones principales, `###` para sub-secciones.
- No saltes niveles (`##` → `####` es incorrecto).

### Listas

- Bullets con `-`, nunca `*` ni `+`.
- Listas numeradas con `1.` (auto-incremento).
- Línea vacía antes y después de toda lista.

### Bloques de código

- Siempre fenced (` ``` `) con lenguaje explícito: `typescript`, `bash`, `json`, `markdown`, etc.
- Código inline con backtick para snippets cortos: `npm run dev`, `NEXT_PUBLIC_API_URL`.

### Tablas

- Pipe delimitado con separador `---`.
- Celdas alineadas, contenido conciso.

```markdown
| Campo       | Tipo     | Descripción              |
|-------------|----------|--------------------------|
| `sessionId` | `string` | ID único de la sesión.   |
| `agentId`   | `string` | ID del agente asignado.  |
```

### Enlaces

- Texto descriptivo, nunca "aquí" o "click here".
- Rutas relativas para archivos del mismo repositorio.

```markdown
Consulta la [guía de despliegue](docs/deploy.md) para más detalles.
```

---

## Idioma / Language

**English is the primary language for all technical documents in this project.**

- Write READMEs, PRPs, specs, docs, and changelogs in English by default, even when the user request is in Spanish.
- If the document explicitly targets end-users in Spanish (e.g. a localized help page), write it in Spanish — but always produce the English version first.
- `apps/web` and `apps/cim` are multilingual (EN/ES) at the UI level via `@webtoq/i18n`; that does not affect the language of technical documentation.

## Estilo de escritura

- **Voz activa** y oraciones cortas.
- **Terminología consistente**: elige un término y no lo mezcles (`feature` vs `característica` → elige uno).
- **Semantic line breaks**: una idea por línea para facilitar diffs en git.
- Sin comentarios editoriales en el archivo entregado — solo contenido útil para el lector.

---

## Flujo de trabajo

Cuando el usuario pida escribir o revisar un `.md`:

1. **Identifica el tipo**: PRP, README, spec, docs, changelog, etc.
2. **Pregunta si falta** el objetivo principal o el público objetivo — no preguntes si ya es claro.
3. **Redacta** aplicando todas las reglas de sintaxis y estilo.
4. **Autovalida** con el checklist antes de entregar.

---

## Checklist de validación

Antes de entregar cualquier `.md`, verifica:

- [ ] Sin `#` en el cuerpo (solo en el título del archivo).
- [ ] Encabezados sin saltos de nivel.
- [ ] Listas usan `-` o `1.` de forma consistente.
- [ ] Bloques de código con lenguaje explícito.
- [ ] Enlaces con texto descriptivo.
- [ ] Tablas bien formateadas.
- [ ] Escritura en voz activa, terminología consistente.
