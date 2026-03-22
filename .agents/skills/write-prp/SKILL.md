---
name: write-prp
description: Redacta un Product Requirement Prompt (PRP) para agentes de codificación, siguiendo Context Engineering y el Framework G3 (Guidelines · Guidance · Guardrails). Úsalo cuando el usuario pida escribir, crear, actualizar o refinar un PRP, especificación de feature, documento de requerimientos o Product Requirement Prompt. Genera el archivo en `PRPs/00X-feature-name.md`.
---

# Write PRP — Product Requirement Prompt

> "No es un simple prompt. Es un documento estructurado que le dice a la IA exactamente qué construir."

Un PRP aplica **Context Engineering en 5 capas**: provee el contexto suficiente para que un agente de codificación (como Claude Code) entienda el sistema, el dominio, la tarea, cómo interactuar y qué producir — sin ambigüedades.

---

## G1 · Guidelines — Estándar de todo PRP

Cada PRP debe cubrir las **5 capas de contexto**, en orden:

| # | Capa | Contenido clave |
|---|------|-----------------|
| 1 | **Sistema** | Arquitectura, stack tecnológico, restricciones de infraestructura |
| 2 | **Dominio** | Reglas de negocio, entidades, flujos y contratos de datos |
| 3 | **Tarea** | Objetivo específico de esta sesión de implementación |
| 4 | **Interacción** | Estilo de comunicación, preguntas, formato de respuesta esperado |
| 5 | **Respuesta** | Estructura del output: archivos, rutas, criterios de aceptación |

**Naming de archivos:** `PRPs/00X-feature-name.md`
- Prefijo numérico de 3 dígitos ordenable: `001-`, `002-`, `003-`…
- Descripción corta en kebab-case: `user-auth`, `chat-session`, `payment-flow`
- Ejemplo: `PRPs/003-anonymous-users.md`

**Header de todo PRP:**
```
# PRP: [Nombre del feature]
> Versión: 1.0 · Fecha: YYYY-MM-DD · Owner: [rol o email] · Estado: Draft | Ready | Done
```

---

## G2 · Guidance — Cómo redactar un PRP paso a paso

### Paso 1 — Recopilar información

Antes de escribir una sola línea del PRP, pregunta al usuario lo que falte:

- **Nombre del feature** (para el nombre de archivo y título)
- **Objetivo de negocio** (1–3 líneas: qué problema resuelve, qué métrica mejora)
- **Usuarios / stakeholders** principales (quién lo usa, quién lo aprueba)
- **Restricciones conocidas** (fecha límite, paquetes afectados, integraciones externas)
- **Draft existente** (si hay un borrador, ruta del archivo)

Si la información es suficiente para escribir el PRP, **no preguntes más** — empieza a redactar.

### Paso 2 — Analizar el contexto del proyecto

Antes de completar las capas, revisa el codebase actual para:

- Identificar paquetes `@webtoq/*` afectados y sus contratos existentes
- Detectar componentes o hooks relevantes ya implementados
- Confirmar patrones arquitectónicos aplicables (App Router, Zustand, TanStack Query, etc.)
- Verificar restricciones del `CLAUDE.md` del proyecto

### Paso 3 — Completar las 5 capas

Usa la **Plantilla de 5 Capas** de la sección siguiente. Completa cada capa con el nivel de detalle adecuado:

- **Sistema y Dominio**: alto nivel — arquitectura, no código
- **Tarea**: preciso — qué construir, qué paquetes tocar, rutas de archivos
- **Respuesta**: testeable — criterios de aceptación verificables, Quality Gates

### Paso 4 — Escribir el archivo

Genera el archivo en `PRPs/00X-feature-name.md` siguiendo la plantilla. Usa el siguiente número disponible en la secuencia.

### Paso 5 — Confirmar con el usuario antes de marcar Ready

El status del PRP recién creado **siempre es `Draft`**. Nunca lo marques como `Ready` al crear. Una vez escrito el PRP, notifica al usuario que el documento está listo para revisión. Solo cambia el status a `Ready` cuando el usuario lo confirme explícitamente (por ejemplo: "aprueba el PRP", "márcalo como ready"). Solo cambia a `Done` cuando la implementación esté completada y todos los Quality Gates pasen.

---

## G2 · 5-Layer Template

```markdown
# PRP: [Feature name]

> Version: 1.0 · Date: YYYY-MM-DD · Owner: [role or email] · Status: Draft | Ready | Done

---

## 1. System

Infrastructure and stack context for this task:

- **Monorepo**: Turborepo + pnpm workspaces
- **Affected apps**: `apps/web` (port 3000) | `apps/cim` (port 3001) | both
- **Involved `@webtoq/*` packages**: list relevant packages
- **Tech stack**: Next.js 14 App Router, TypeScript strict, Tailwind (no CSS files), Shadcn/ui
- **Infrastructure constraints**: SST/AWS, bundle limits, required environment variables
- **Dependency hierarchy to respect**: types → domain/contracts → api → api-react → chat-core/chat-engine → ui/widget → apps

---

## 2. Domain

Business rules, entities, and relevant flows:

- **Main entities**: list involved domain entities
- **Business rules**: listed as FR-01, FR-02… (functional requirements)
- **Key flows**: description of user or system flows (may include pseudo-diagram)
- **Data contracts**: schemas in `@webtoq/contracts`, models in `@webtoq/domain`, affected mappers
- **WebSocket events**: if applicable, events emitted/listened via `@webtoq/transport`
- **Non-functional constraints**: performance, security, accessibility (WCAG), scalability

---

## 3. Task

Specific objective for this implementation session:

- **Objective**: 1–3 line description of what is being implemented
- **Files to create**: exact paths (e.g. `packages/contracts/src/schemas/session.ts`)
- **Files to modify**: paths and expected change
- **Out of bounds**: what NOT to touch in this session
- **User stories**:
  - US-01: As a [role], I want [objective] so that [benefit].
  - US-02: …
- **Acceptance criteria**:
  - AC-01: Given [context], when [action], then [expected result].
  - AC-02: …

---

## 4. Interaction

How Claude should behave during implementation:

- **Tone**: technical-direct | collaborative-explanatory | code-only
- **Allowed questions**: if ambiguity is found in X, ask before proceeding
- **Intermediate response format**: brief comments per step | detailed explanations | silence (code only)
- **Language**: English (default — all technical artifacts in this project are English-first)
- **If critical info is missing**: pause and ask; never invent business rules

---

## 5. Response

Expected output structure at the end of the session:

- **Code artifacts**: list of generated/modified files with brief description
- **Quality Gates** (all must pass before considering complete):
  - [ ] `pnpm build` without errors
  - [ ] `pnpm lint` without errors (TypeScript strict, no `any`)
  - [ ] Relevant tests pass (if they exist)
  - [ ] Manual smoke test: description of the flow to validate manually
- **Documentation to update**: `/docs`, `/PRPs`, package changelog if applicable
- **Do not include**: new unit tests (unless explicitly requested), unrelated global style changes, out-of-scope refactors
```

---

## G3 · Guardrails — Lo que nunca debes hacer

**Sobre la información:**
- **Nunca inventes** reglas de negocio, flujos o contratos de datos no confirmados — pregunta.
- **Nunca asumas** qué paquetes tocar si no lo confirmó el usuario — verifica el codebase.
- **Nunca omitas** la capa Sistema si el PRP es para Claude Code — es el contexto más crítico.

**Sobre la estructura:**
- **Nunca cambies** el orden de las 5 capas sin que el usuario lo pida explícitamente.
- **Nunca mezcles** niveles de abstracción: dominio es negocio, tarea es técnico, no al revés.
- **Nunca escribas** snippets de código completos en el PRP — solo rutas, patrones y contratos.

**Sobre el naming:**
- **Nunca** nombres un archivo `prp.md`, `PRP_v2.md`, `feature-spec.md` — siempre `00X-kebab-case.md`.
- **Nunca** reutilices un número de prefijo existente — revisa los archivos en `PRPs/` antes.

**Sobre el scope:**
- **Nunca** incluyas tareas de múltiples features en un solo PRP — un PRP, un feature.
- **Nunca** marques el PRP como `Status: Ready` al momento de crearlo — el status inicial es siempre `Status: Draft`.
- **Nunca** cambies el status a `Ready` sin confirmación explícita del usuario.
- **Nunca** marques `Ready` si algún criterio de aceptación está vacío o ambiguo.

**Sobre el idioma:**
- **PRPs siempre en inglés** — es el idioma principal de todos los artefactos técnicos del proyecto.
- **Nunca redactes** secciones del PRP en español; si el usuario pide algo en español, produce el documento en inglés de todas formas.
- **La capa de Interacción** puede especificar `Language: Spanish` solo si el usuario lo solicita explícitamente; el valor por defecto es siempre `English`.

**Sobre el codebase:**
- **Respetar** la jerarquía de dependencias del monorepo (types → … → apps). Nunca invertirla.
- **Respetar** que hay dos SST configs (`sst.config.ts` para web, `sst-cim.config.ts` para cim).
- **No agregar** CSS files — solo Tailwind. No agregar framework de tests sin alineación del equipo.

---

## Ejemplo de interacción

**Usuario:** Quiero un PRP para agregar usuarios anónimos al widget de chat.

**Claude:**
1. Verifica si ya existe un PRP relacionado en `PRPs/`.
2. Pregunta (si falta): ¿qué datos debe persistir el usuario anónimo? ¿hay sesión en backend?
3. Revisa `@webtoq/transport`, `@webtoq/chat-engine` y `apps/web/stores/` para contexto.
4. Genera `PRPs/004-anonymous-users.md` con las 5 capas completas.
5. Confirma con el usuario antes de marcar `Estado: Ready`.
