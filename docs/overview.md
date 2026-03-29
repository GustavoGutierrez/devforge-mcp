# DevForge MCP — Visión General del Proyecto

> Documento de referencia para developers que quieren entender, usar y extender DevForge MCP.
> Fecha de generación: 2026-03-22.

---

## Tabla de contenidos

1. [Qué es DevForge MCP](#1-qué-es-devforge-mcp)
2. [Herramientas MCP implementadas](#2-herramientas-mcp-implementadas)
3. [Casos de uso](#3-casos-de-uso)
4. [Arquitectura](#4-arquitectura)
5. [Qué está en extensión o en desarrollo](#5-qué-está-en-extensión-o-en-desarrollo)
6. [Cómo extender el proyecto](#6-cómo-extender-el-proyecto)

---

## 1. Qué es DevForge MCP

DevForge MCP es un **servidor MCP** (Model Context Protocol) escrito en Go que actúa como una capa de inteligencia transversal al ciclo de desarrollo de software. Expone herramientas especializadas para UI/diseño, procesamiento de imágenes y gestión de sistemas de diseño a través del **transporte stdio** de MCP, y viene acompañado de una **CLI/TUI** interactiva construida con Bubble Tea.

### El problema que resuelve

El desarrollo de interfaces modernas implica decisiones repetitivas y propensas a la inconsistencia: ¿qué tokens usar?, ¿cómo estructurar un layout en Next.js vs Astro?, ¿cómo optimizar imágenes para producción?. Cada proyecto rehace estas decisiones desde cero, acumulando deuda de diseño y falta de coherencia.

DevForge MCP centraliza ese conocimiento en una capa de herramientas invocables por agentes de IA o por el developer directamente, garantizando:

- **Consistencia estructural** — los mismos patrones y convenciones en cualquier módulo o proyecto.
- **Diseño replicable** — paletas, tokens y componentes ready para producción.
- **Aceleración del flujo** — auditorías, generación de layouts y optimización de assets en segundos.
- **Calidad integrada** — validaciones de accesibilidad, tipografía y responsividad en el propio flujo.

### Contexto técnico

- **Transporte**: stdio exclusivamente (JSON-RPC 2.0). El cliente MCP lanza el binario como proceso hijo; los mensajes fluyen por stdin/stdout.
- **Base de datos**: SQLite/libSQL con FTS5 para búsqueda full-text y vector ANN (via libsql_vector_idx) para búsqueda semántica cuando Ollama está disponible.
- **Procesamiento multimedia**: motor Rust de alto rendimiento (`dpf` / DevPixelForge) invocado desde Go mediante un bridge con `StreamClient`. Soporta imágenes, video y audio. Requiere FFmpeg para video/audio.
- **Generación de imágenes IA**: integración con Google Gemini API (`gemini-2.5-flash-image`).
- **Lenguaje**: Go 1.24+ con CGO habilitado (requerido por `go-libsql`).

### Stacks frontend soportados

| Framework | CSS Mode |
|-----------|----------|
| SPA Vite 8 (`spa-vite`) | Tailwind CSS v4+ (`tailwind-v4`) |
| Astro (`astro`) | CSS moderno con custom properties (`plain-css`) |
| Next.js (`next`) | |
| SvelteKit (`sveltekit`) | |
| Nuxt.js (`nuxt`) | |
| Vanilla (`vanilla`) | |

**Tailwind v4**: sin `tailwind.config.js`. Tokens en CSS nativo (`@theme`, `:root`, `@property`, `@layer`). Importación vía `@import "tailwindcss";`.

---

## 2. Herramientas MCP implementadas

Cada tool devuelve siempre JSON estructurado. Los errores se devuelven como `{"error": "mensaje"}` — el servidor nunca hace panic en un handler de tool.

### Convención de stack

Las tools de layout y diseño requieren el objeto `stack`:

```json
{
  "stack": {
    "css_mode": "tailwind-v4",
    "framework": "next"
  }
}
```

---

### `analyze_layout`

**Propósito**: Audita markup HTML/JSX para detectar problemas de layout.

**Fichero**: `internal/tools/analyze_layout.go`

**Qué detecta**:
- Imágenes sin atributo `alt` (error de accesibilidad).
- Uso de `<div>` sin roles ARIA ni HTML semántico.
- Tamaños de fuente en píxeles en lugar de `rem`/`em`.
- Uso de `tailwind.config.js` en proyectos Tailwind v4 (incorrecto).
- `@apply` fuera de un bloque `@layer` en Tailwind v4.
- Clases Tailwind detectadas en proyectos `plain-css` (inconsistencia).
- Ausencia de media queries o breakpoints en layouts mobile.
- Espaciados `0px` en lugar de `0` (shorthand correcto).

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `markup` | string (required) | HTML/JSX/Svelte/Vue a analizar |
| `stack` | object (required) | `css_mode` + `framework` |
| `page_type` | string | `landing`, `dashboard`, `form` |
| `device_focus` | string | `mobile`, `desktop`, `both` |

**Output**:

```json
{
  "summary": "Found 2 errors, 1 warning in next/tailwind-v4 layout.",
  "issues": [
    {
      "severity": "error",
      "category": "accessibility",
      "description": "One or more <img> elements are missing alt attributes.",
      "suggestion": "Add descriptive alt text to all images."
    }
  ],
  "score": 77
}
```

Los resultados se persisten automáticamente en la tabla `audits` de la base de datos.

---

### `suggest_layout`

**Propósito**: Genera scaffolds de layout listos para copiar en proyectos reales.

**Fichero**: `internal/tools/suggest_layout.go`

El generador produce snippets adaptados a cada framework y CSS mode. Para `tailwind-v4` genera tokens en `@theme { }` y clases de utilidad; para `plain-css` genera custom properties en `:root` y class names con scoping.

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `description` | string (required) | Descripción del layout deseado |
| `stack` | object (required) | `css_mode` + `framework` |
| `fidelity` | string (required) | `wireframe`, `mid`, `production` |
| `tokens_profile` | object | Tokens existentes a incorporar |

**Output**:

```json
{
  "layout_name": "SaasDashboardLayout",
  "files": [
    { "path": "app/components/SaasDashboardLayout.tsx", "snippet": "..." }
  ],
  "css_snippets": [
    { "path": "app/globals.css", "snippet": "..." }
  ],
  "rationale": "Generated a mid fidelity layout for next using tailwind-v4. CSS tokens are defined in @theme layer."
}
```

Frameworks con soporte específico: `astro` (`.astro`), `next` (`.tsx`), `sveltekit` (`.svelte`), `nuxt` (`.vue`), `spa-vite` (`.tsx`), `vanilla` (`.html`).

---

### `manage_tokens`

**Propósito**: Gestiona design tokens (colores, espaciados, tipografía) en la base de datos.

**Fichero**: `internal/tools/manage_tokens.go`

Tres modos de operación:
- **`read`**: devuelve los tokens actuales (desde DB o defaults integrados).
- **`plan-update`**: calcula el diff entre tokens actuales y una propuesta, sin persistir.
- **`apply-update`**: calcula el diff y persiste los cambios en la tabla `tokens`.

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `mode` | string (required) | `read`, `plan-update`, `apply-update` |
| `css_mode` | string (required) | `tailwind-v4`, `plain-css` |
| `scope` | string (required) | `colors`, `spacing`, `typography`, `all` |
| `proposal` | object | Mapa `{key: value}` de tokens a aplicar |

**Output**:

```json
{
  "current_tokens": { "--color-primary": "#3b82f6" },
  "diff": { "--color-primary": { "old": "#3b82f6", "new": "#0ea5e9" } },
  "instructions": "Apply tokens in @theme { } block in your global CSS file:\n\n@theme {\n  ..."
}
```

Tokens por defecto integrados (fallback cuando la DB está vacía): colores semánticos (primary, secondary, background, surface, text, muted), escala de espaciados (xs→xl) y escala tipográfica (sm→4xl).

---

### `store_pattern`

**Propósito**: Persiste un patrón de UI en la base de datos para reutilización futura.

**Fichero**: `internal/tools/store_pattern.go`

Además de insertar en la tabla `patterns`, indexa automáticamente en la tabla virtual FTS5 (`patterns_fts`) y, si Ollama está disponible, genera el embedding vectorial en background (goroutine no bloqueante).

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `name` | string (required) | Nombre del patrón |
| `framework` | string (required) | Framework objetivo |
| `css_mode` | string (required) | CSS mode |
| `snippet` | string (required) | Código HTML/JSX/Svelte |
| `category` | string | `landing`, `dashboard`, `form`, `component`, `other` |
| `domain` | string | `frontend`, `backend`, `fullstack`, `devops`, `any` |
| `tags` | string | Tags separadas por comas |
| `css_snippet` | string | CSS asociado (opcional) |
| `description` | string | Descripción del patrón |

**Output**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Hero Split",
  "created_at": "2026-03-22T10:00:00Z"
}
```

---

### `list_patterns`

**Propósito**: Consulta patrones almacenados con filtros, búsqueda full-text o búsqueda semántica.

**Fichero**: `internal/tools/list_patterns.go`

Tres modos de búsqueda (con detección automática):

| Modo | Activación | Mecanismo |
|------|-----------|-----------|
| `filter` | Sin `query` | `SQL WHERE` con filtros exactos |
| `fts` | Con `query` | FTS5 full-text search sobre name, tags, description |
| `semantic` | Con `query` + Ollama disponible | Vector ANN con fallback a FTS5 |

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `query` | string | Búsqueda por keywords o lenguaje natural |
| `mode` | string | `fts`, `semantic`, `filter` (auto si se omite) |
| `domain` | string | Filtro por dominio |
| `framework` | string | Filtro por framework |
| `css_mode` | string | Filtro por CSS mode |
| `limit` | int | Máx. resultados (default 20) |

---

### `suggest_color_palettes`

**Propósito**: Genera propuestas de paletas de color coherentes para distintos tipos de aplicación.

**Fichero**: `internal/tools/suggest_color_palettes.go`

Biblioteca interna de 8 paletas predefinidas con scoring por relevancia (use case + mood + brand keywords). Cada paleta incluye 7 tokens semánticos: `background`, `surface`, `primary`, `primary-soft`, `accent`, `text`, `muted`.

**Paletas disponibles**: Fintech Calm Blue, SaaS Indigo, Marketing Vibrant, Minimal Light, Dark Premium, E-commerce Warm, Health & Wellness, Developer Tools.

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `use_case` | string (required) | Ej.: `"SaaS dashboard"`, `"marketing site"` |
| `mood` | string | Ej.: `"calm"`, `"bold"`, `"minimal"`, `"professional"` |
| `brand_keywords` | string[] | Keywords de marca para afinar el scoring |
| `count` | int | Número de paletas a devolver (default 3) |

**Output**:

```json
{
  "palettes": [
    {
      "name": "SaaS Indigo",
      "description": "Modern indigo palette for SaaS dashboards...",
      "tokens": {
        "background": "#ffffff",
        "surface": "#f8fafc",
        "primary": "#6366f1",
        "primary-soft": "#eef2ff",
        "accent": "#f59e0b",
        "text": "#1e293b",
        "muted": "#64748b"
      }
    }
  ]
}
```

---

### `generate_ui_image`

**Propósito**: Genera imágenes de UI (wireframes, mockups, ilustraciones) usando Google Gemini.

**Fichero**: `internal/tools/generate_ui_image.go`

**Requiere**: `gemini_api_key` configurada (via `configure_gemini` o en `~/.config/devforge/config.json`).

Modelo predeterminado: `gemini-2.5-flash-image`. Configurable via campo `image_model` en el archivo de config.

Los datos de imagen se reciben directamente como bytes (no base64) desde la API de Gemini y se guardan en el `output_path` especificado. El directorio se crea automáticamente si no existe.

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `prompt` | string (required) | Descripción de la imagen a generar |
| `style` | string (required) | `wireframe`, `mockup`, `illustration` |
| `output_path` | string (required) | Ruta donde guardar la imagen |
| `width` | int | Ancho en píxeles (default 1280) |
| `height` | int | Alto en píxeles (default 720) |

**Output**:

```json
{
  "path": "assets/generated/hero.png",
  "width": 1280,
  "height": 720,
  "prompt_used": "Create a high-fidelity UI mockup of: SaaS dashboard..."
}
```

---

### `optimize_images`

**Propósito**: Comprime imágenes PNG/JPEG y genera variantes WebP/AVIF.

**Fichero**: `internal/tools/optimize_images.go`

**Requiere**: `bin/dpf` ejecutable.

Invoca el motor Rust via `StreamClient` (proceso persistente, overhead ~5ms por request). Soporta múltiples imágenes en la misma llamada con paralelismo configurable.

**Input**:

```json
{
  "inputs": [
    {
      "path": "assets/raw/hero.png",
      "max_width": 1920,
      "max_height": 1080,
      "formats": ["webp", "avif"],
      "quality": 80
    }
  ],
  "parallelism": 4
}
```

**Output**:

```json
{
  "results": [
    {
      "source_path": "assets/raw/hero.png",
      "outputs": [
        {
          "format": "webp",
          "path": "assets/optimized/hero.webp",
          "width": 1920,
          "height": 1080,
          "approx_size_kb": 120
        }
      ]
    }
  ]
}
```

---

### `generate_favicon`

**Propósito**: Genera un pack completo de favicons desde una imagen fuente SVG/PNG.

**Fichero**: `internal/tools/generate_favicon.go`

**Requiere**: `bin/dpf` ejecutable.

Genera variantes en ICO, PNG y SVG para todos los tamaños estándar: 16, 32, 48 (browser), 180 (Apple Touch Icon), 192, 512 (Android/PWA). Incluye `web.manifest` y snippets HTML listos para pegar.

La imagen final es siempre cuadrada — si la fuente no lo es, aplica letterboxing (nunca stretching).

**Input**:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `source_path` | string (required) | Ruta de la imagen fuente (PNG o SVG) |
| `background_color` | string | Color de fondo hex (default `#ffffff`) |
| `sizes` | int[] | Tamaños en px (default `[16,32,48,180,192,512]`) |
| `formats` | string[] | `ico`, `png`, `svg` (default todos) |

**Output**:

```json
{
  "icons": [
    { "size": 32, "format": "png", "path": "favicons/favicon-32x32.png" }
  ],
  "html_snippets": [
    "<link rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"/favicon-32x32.png\">"
  ]
}
```

---

### `image_resize`

**Propósito**: Redimensiona imágenes a múltiples anchos o por porcentaje.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output_dir": "public/img",
  "widths": [320, 640, 960, 1280],
  "format": "webp",
  "quality": 85,
  "linear_rgb": true
}
```

**Output**:

```json
{
  "success": true,
  "variants": [
    {"path": "public/img/photo-320.webp", "width": 320, "height": 213, "format": "webp", "size_kb": 18},
    {"path": "public/img/photo-640.webp", "width": 640, "height": 427, "format": "webp", "size_kb": 45}
  ],
  "elapsed_ms": 120
}
```

---

### `image_crop`

**Propósito**: Recorta una imagen a dimensiones específicas.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output": "public/img/cropped.jpg",
  "x": 100,
  "y": 50,
  "width": 800,
  "height": 600
}
```

---

### `image_rotate`

**Propósito**: Rota y/o voltea una imagen.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output": "public/img/rotated.jpg",
  "angle": 90,
  "flip_h": false,
  "flip_v": true
}
```

---

### `image_watermark`

**Propósito**: Añade una marca de agua (texto o imagen) a una imagen.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output": "public/img/watermarked.jpg",
  "text": "© 2026 Company",
  "position": "center",
  "opacity": 0.5,
  "color": "#ffffff",
  "size": 48
}
```

---

### `image_adjust`

**Propósito**: Ajusta propiedades de la imagen (brillo, contraste, saturación, blur, sharpen).

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output": "public/img/adjusted.jpg",
  "brightness": 10,
  "contrast": 15,
  "saturation": -10,
  "blur": 0,
  "sharpen": 1.5
}
```

---

### `image_quality`

**Propósito**: Optimiza calidad de imagen a un tamaño objetivo usando búsqueda binaria.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output": "public/img/optimized.jpg",
  "target_size_kb": 50,
  "format": "webp",
  "max_quality": 95,
  "min_quality": 30
}
```

---

### `image_srcset`

**Propósito**: Genera variantes responsivas para el atributo srcset.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/hero.jpg",
  "output_dir": "public/img",
  "widths": [320, 640, 960, 1280, 1920],
  "sizes": ["100vw", "(min-width: 768px) 50vw"],
  "format": "webp"
}
```

---

### `image_exif`

**Propósito**: Operaciones EXIF (strip, preserve, extract, auto-orient).

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output": "public/img/stripped.jpg",
  "exif_op": "strip"
}
```

Operaciones disponibles:
- `strip`: Elimina todos los metadatos EXIF
- `preserve`: Copia la imagen sin modificar EXIF
- `extract`: Devuelve los datos EXIF en el campo `data`
- `auto_orient`: Corrige automáticamente la orientación de la imagen

---

### `image_convert`

**Propósito**: Convierte una imagen a otro formato.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.png",
  "output": "public/img/photo.webp",
  "format": "webp",
  "quality": 85,
  "width": 800,
  "height": 600
}
```

---

### `image_placeholder`

**Propósito**: Genera placeholders de imagen (LQIP, color dominante, gradiente CSS).

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output": "public/img/placeholder.webp",
  "kind": "lqip",
  "lqip_width": 20,
  "inline": true
}
```

**Output**:

```json
{
  "success": true,
  "output_path": "public/img/placeholder.webp",
  "data_base64": "data:image/webp;base64,...",
  "dominant_color": "#4a90d9",
  "css_gradient": "linear-gradient(#4a90d9, #2d5a87)"
}
```

---

### `image_palette`

**Propósito**: Reduce la paleta de colores o extrae colores dominantes.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "input": "assets/photo.jpg",
  "output_dir": "public/img",
  "max_colors": 16,
  "dithering": 0.5,
  "format": "gif"
}
```

---

### `image_sprite`

**Propósito**: Genera un sprite sheet desde múltiples imágenes con CSS opcional.

**Fichero**: `internal/tools/image_tools.go`

**Requiere**: `bin/dpf` ejecutable.

**Input**:

```json
{
  "inputs": ["assets/icon1.png", "assets/icon2.png", "assets/icon3.png"],
  "output": "public/img/sprites.png",
  "cell_size": 32,
  "columns": 3,
  "padding": 4,
  "generate_css": true
}
```

---

### `configure_gemini`

**Propósito**: Guarda el API key de Gemini y lo recarga en caliente sin reiniciar el servidor.

**Fichero**: `internal/tools/configure_gemini.go`

Hot-reload implementado mediante callback `updateConfig(key string)` pasado al handler. El `mcpApp` en el entry point actualiza `geminiKey` detrás de un `sync.RWMutex`.

**Input**:

```json
{ "api_key": "AIzaXXXXXXXXXXX" }
```

**Output**:

```json
{
  "config_path": "/home/user/.config/devforge/config.json",
  "status": "saved"
}
```

---

### `store_architecture` *(reciente, no registrada en MCP aún)*

**Propósito**: Persiste descripciones de arquitecturas de software (ADRs, decisiones técnicas) en la DB.

**Fichero**: `internal/tools/store_architecture.go`

Misma estructura que `store_pattern` pero orientada a decisiones de arquitectura fullstack/backend. Indexa en `architectures_fts` y genera embeddings en background si Ollama está disponible.

> **Nota**: Este tool está implementado en `internal/tools/` pero no está registrado en el servidor MCP en `cmd/devforge-mcp/main.go` al momento de la revisión. Es una extensión en progreso.

---

## 3. Casos de uso

### 3.1 Agente de auditoría de frontend

Un agente de IA recibe markup de una página existente y usa `analyze_layout` para obtener un informe estructurado de issues de accesibilidad, tipografía y consistencia de tokens. Luego consulta `list_patterns` para ver si hay patrones similares almacenados que sirvan de referencia para proponer mejoras.

```
analyze_layout → list_patterns (buscar referencia) → sugerir cambios
```

### 3.2 Generación de un nuevo componente de UI

Un developer describe el componente que necesita ("hero section con imagen a la derecha y CTA dual"). `suggest_layout` devuelve el código listo para el framework elegido (Next.js, Astro, SvelteKit...) con sus tokens CSS. Si el resultado es bueno, `store_pattern` lo guarda para reutilización.

```
suggest_layout → revisar resultado → store_pattern
```

### 3.3 Inicialización del sistema de diseño de un proyecto

Al comenzar un proyecto nuevo, `suggest_color_palettes` genera opciones de paleta por tipo de aplicación. `manage_tokens` (modo `apply-update`) persiste los tokens elegidos en la DB. `suggest_layout` los consume para generar layouts coherentes desde el inicio.

```
suggest_color_palettes → manage_tokens (apply) → suggest_layout
```

### 3.4 Optimización de assets para producción

Un pipeline CI/CD o un agente `asset-optimizer` pasa las imágenes de un proyecto a `optimize_images` para comprimir y generar WebP/AVIF, y a `generate_favicon` para crear el pack de iconos completo con sus snippets HTML listos para el `<head>`.

```
optimize_images → generate_favicon → copiar html_snippets al template
```

### 3.5 Prototipado visual con IA

Un agente de ideación usa `generate_ui_image` para materializar visualmente un concepto de pantalla antes de escribir código. El resultado (mockup 1280×720) sirve de referencia para el developer. Si la imagen es demasiado pesada, `optimize_images` la convierte a WebP para uso en documentación o Storybook.

```
configure_gemini → generate_ui_image → optimize_images
```

### 3.6 Exploración semántica de la biblioteca de patrones

Con Ollama disponible, un agente puede buscar patrones por intención en lugar de keywords: `list_patterns` con `mode: "semantic"` y `query: "navigation that collapses on mobile"` encuentra patrones similares por embedding vectorial.

```
list_patterns (semantic) → recuperar snippet → adaptar al proyecto
```

---

## 4. Arquitectura

### 4.1 Diagrama de componentes

```
┌─────────────────────────────────────────────────────────────┐
│                       MCP Client                            │
│           (Claude Code, Claude Desktop, VS Code,             │
│            Cursor, OpenCode, cualquier cliente stdio)        │
└──────────────────────┬──────────────────────────────────────┘
                       │ JSON-RPC 2.0 / stdio
                       ▼
┌─────────────────────────────────────────────────────────────┐
│               cmd/devforge-mcp/main.go                     │
│                                                             │
│   mcpApp {                                                  │
│     srv: *tools.Server     ← handlers de tools              │
│     geminiKey, imageModel  ← hot-reloadable via RWMutex     │
│   }                                                         │
│                                                             │
│   Registro de tools → mcp-go SDK → ServeStdio()             │
└──────┬─────────────────────┬───────────────────────────────┘
       │                     │
       ▼                     ▼
┌─────────────┐    ┌──────────────────────────┐
│ internal/db │    │ internal/dpf          │
│             │    │ StreamClient             │
│ SQLite/     │    │ (proceso Rust persistente) │
│ libSQL      │    └──────────┬───────────────┘
│ FTS5 + vec  │               │ stdin/stdout JSON
│             │               ▼
│ tables:     │    ┌──────────────────────────┐
│  patterns   │    │ bin/dpf      │
│  architectures   │ (motor Rust de imágenes) │
│  tokens     │    └──────────────────────────┘
│  audits     │
│  assets     │    ┌──────────────────────────┐
│  palettes   │    │ Google Gemini API        │
└─────────────┘    │ (generate_ui_image)      │
                   └──────────────────────────┘

       ┌──────────────────────────────────────┐
       │          cmd/devforge/main.go        │
       │          CLI/TUI (Bubble Tea)         │
       │                                      │
       │  Mismas dependencias: DB + tools.Server │
       │  Vistas: Home, Browse, Analyze,      │
       │  Generate, Settings, MCP Setup, etc. │
       └──────────────────────────────────────┘
```

### 4.2 Servidor MCP (`cmd/devforge-mcp`)

El entry point inicializa los componentes en orden:

1. **Config** (`internal/config`) — lee `~/.config/devforge/config.json` o `DEV_FORGE_CONFIG`.
2. **DB** (`internal/db`) — abre la base de datos (path relativo al binario), ejecuta migraciones, activa WAL mode y FKs.
3. **Embedder** (`internal/db.EmbeddingClient`) — si Ollama responde en `ollama_url`, se activa búsqueda semántica; si no, degrada a FTS5 silenciosamente.
4. **StreamClient** (`internal/dpf`) — lanza el proceso Rust `devforge-imgproc` como proceso hijo persistente. Si el binario no está, las tools de imagen devuelven error estructurado sin crashear el servidor.
5. **tools.Server** — struct con `DB`, `Imgproc`, `Embedder` y `GetConfig` compartido.
6. **Backfill** — goroutine en background que genera embeddings para filas con `embedding IS NULL`.
7. **MCP server** — registra los 10 tools via `mcp-go` SDK y sirve via `ServeStdio()`.

El hot-reload de Gemini API key funciona así: `configure_gemini` llama `app.setGeminiKey(key)` que actualiza `mcpApp.geminiKey` bajo `sync.RWMutex`; `generate_ui_image` siempre lee `app.getGeminiKey()` en cada invocación.

### 4.3 Base de datos (`internal/db`)

**Driver**: `github.com/tursodatabase/go-libsql` (CGO required). Compatible con SQLite estándar y libSQL con extensiones vectoriales.

**Schema**:

| Tabla | Propósito | Búsqueda |
|-------|-----------|----------|
| `patterns` | Patrones de UI (snippets + metadata) | FTS5 + Vector ANN |
| `architectures` | Decisiones de arquitectura (ADRs) | FTS5 + Vector ANN |
| `tokens` | Design tokens por `css_mode` y `scope` | SQL WHERE |
| `audits` | Informes de análisis de layout | SQL WHERE |
| `assets` | Registro de assets procesados | SQL WHERE |
| `palettes` | Paletas de color guardadas | SQL WHERE |

**Tablas virtuales FTS5** (`patterns_fts`, `architectures_fts`): indexan name, category, tags, description. Permiten operadores booleanos y búsqueda por prefijo.

**Vector ANN** (`patterns_vec_idx`, `architectures_vec_idx`): índices sobre columna `F32_BLOB(768)`. Solo activos si libSQL tiene la extensión `libsql_vector_idx`. Se consultan via `vector_top_k(...)`.

**Embeddings**: generados por `nomic-embed-text` via Ollama (768 dimensiones). Opcionales — si Ollama no está disponible, la fila se guarda con `embedding = NULL` y se busca por FTS5.

**Estrategia de búsqueda**:

| Caso | Mecanismo |
|------|-----------|
| Keywords / tags exactos | FTS5 |
| Filtros por framework, css_mode | SQL WHERE |
| Similitud semántica | Vector ANN (requiere Ollama) |
| Sin query | SQL filter simple |

### 4.4 Bridge imgproc (`internal/dpf`)

Dos clientes disponibles:

- **`StreamClient`** (recomendado para el servidor MCP): mantiene el proceso Rust vivo durante toda la sesión. Goroutine-safe. Ahorra ~5ms de arranque por operación.
- **`Client`** (one-shot): lanza y termina el proceso por cada job. Para uso puntual.

**Job types** disponibles en el protocolo:

| Job | `operation` | Uso |
|-----|-------------|-----|
| `OptimizeJob` | `"optimize"` | Comprimir PNG/JPEG + generar WebP |
| `ResizeJob` | `"resize"` | Variantes responsivas por ancho |
| `ConvertJob` | `"convert"` | Conversión de formato (SVG→WebP, PNG→AVIF) |
| `FaviconJob` | `"favicon"` | Pack completo de favicons |
| `SpriteJob` | `"sprite"` | Sprite sheet + CSS |
| `PlaceholderJob` | `"placeholder"` | LQIP, color dominante, CSS gradient |
| `BatchJob` | `"batch"` | Múltiples operaciones en paralelo |

La comunicación es JSON por stdin/stdout del proceso Rust. El binario pre-compilado vive en `bin/dpf`.

### 4.5 CLI/TUI (`cmd/devforge`)

Construida con Bubble Tea. El `Model` raíz gestiona la navegación entre vistas via `NavigateTo` messages. Cada vista es un sub-model independiente con `Init()`, `Update()`, `View()`.

**Vistas disponibles**:

| View | Descripción |
|------|-------------|
| `ViewHome` | Menú principal con todas las opciones |
| `ViewBrowsePatterns` | Lista de patrones con filtros y búsqueda |
| `ViewBrowseArchitectures` | Lista de arquitecturas con filtros |
| `ViewAnalyzeLayout` | Formulario para auditar un archivo de layout |
| `ViewGenerateLayout` | Formulario para generar layouts |
| `ViewGenerateImages` | Generación de imágenes via Gemini |
| `ViewOptimizeImages` | Optimización de imágenes via Rust |
| `ViewGenerateFavicon` | Generación de pack de favicons |
| `ViewColorPalettes` | Exploración de paletas de color |
| `ViewSettings` | Configuración (Gemini API key, etc.) |
| `ViewMCPSetup` | Asistente para configurar el MCP en IDEs |
| `ViewAddRecord` | Wizard para añadir patrones o arquitecturas |

**Detección automática de stack**: el TUI puede detectar el framework del proyecto leyendo `package.json`, `astro.config.*`, `next.config.*`, `svelte.config.*`, `nuxt.config.*` y prerellenar `framework` y `css_mode` en los formularios.

**MCP Setup wizard**: permite configurar el servidor MCP en OpenCode, Claude Code y VS Code directamente desde la TUI, escribiendo el JSON de configuración en el archivo correcto (global o project-local).

### 4.6 Configuración (`internal/config`)

```go
type Config struct {
    GeminiAPIKey   string `json:"gemini_api_key"`
    OllamaURL      string `json:"ollama_url"`       // default: http://localhost:11434
    EmbeddingModel string `json:"embedding_model"`  // default: nomic-embed-text
    ImageModel     string `json:"image_model"`      // default: gemini-2.5-flash-image
}
```

- **Path**: `~/.config/devforge/config.json` (override via `DEV_FORGE_CONFIG`).
- **Permisos**: `0600` (solo propietario).
- **Compartida** entre el MCP server y la CLI/TUI.
- **Hot-reload**: `configure_gemini` actualiza la clave en memoria sin reiniciar.

---

## 5. Qué está en extensión o en desarrollo

### 5.1 `store_architecture` — implementado, pendiente de registrar en MCP

El handler `StoreArchitecture` en `internal/tools/store_architecture.go` está completamente implementado (inserción en DB, indexado FTS5, embedding en background), pero **no está registrado** en `cmd/devforge-mcp/main.go`. La extensión es mínima: añadir un bloque `s.AddTool(...)` en `registerTools()`.

### 5.2 Búsqueda semántica (vector ANN) — implementada, requiere Ollama

La infraestructura completa está en su lugar:
- Columna `F32_BLOB(768)` en `patterns` y `architectures`.
- Índice `patterns_vec_idx` / `architectures_vec_idx` via `libsql_vector_idx`.
- `EmbeddingClient` que llama a Ollama.
- `list_patterns` con modo `semantic` y fallback a FTS5.
- Backfill de embeddings al arrancar el servidor.

Para activarla: instalar Ollama, hacer `ollama pull nomic-embed-text` y asegurarse de que `ollama_url` en config apunta al endpoint correcto.

### 5.3 libSQL para vector search — opcional, en evaluación

El PRP menciona la posibilidad de migrar a libSQL (Turso embebido) para habilitar vector ANN con `libsql_vector_idx` de forma más robusta, manteniendo compatibilidad 100% local sin dependencia de Turso Cloud. El código ya usa `go-libsql` como driver, por lo que la migración es de configuración, no de código.

### 5.4 Herramientas adicionales planificadas

El PRP-001 describe la intención de extender las tools más allá de UI/diseño hacia otros dominios del stack de desarrollo. La arquitectura está diseñada para esto: agregar un nuevo tool es añadir un fichero en `internal/tools/` y un bloque de registro en `main.go`.

### 5.5 Seeds de la base de datos

El sistema de seeds (`db/seeds/*.sql`, aplicados por `make db-seed`) está preparado pero los ficheros de seed con patrones, paletas y arquitecturas predefinidas están en construcción.

---

## 6. Cómo extender el proyecto

### 6.1 Añadir un nuevo tool MCP

**Paso 1 — Crear el fichero del handler**

Crear `internal/tools/mi_tool.go`:

```go
package tools

import (
    "context"
    "strings"
)

// MiToolInput es el schema de entrada del tool.
type MiToolInput struct {
    Param1 string `json:"param1"`
    Stack  StackMeta `json:"stack,omitempty"`
}

// MiToolOutput es el schema de salida.
type MiToolOutput struct {
    Result string `json:"result"`
}

// MiTool implementa el MCP tool mi_tool.
func (s *Server) MiTool(ctx context.Context, input MiToolInput) string {
    if strings.TrimSpace(input.Param1) == "" {
        return errorJSON("param1 is required")
    }

    // ... lógica ...

    return mustJSON(MiToolOutput{Result: "..."})
}
```

**Convenciones**:
- Un fichero por tool.
- Los errores siempre vía `errorJSON("mensaje")`, nunca `panic`.
- Los tipos input/output deben ser structs exportados con json tags.
- Si el tool necesita DB, usar `s.DB` con `context.Context`. Si necesita dpf, usar `s.DPF`.

**Paso 2 — Registrar en el servidor**

En `cmd/devforge-mcp/main.go`, dentro de `registerTools(s, app)`:

```go
s.AddTool(mcp.NewTool("mi_tool",
    mcp.WithDescription("Descripción clara del tool."),
    mcp.WithString("param1", mcp.Required(), mcp.Description("Descripción del param1")),
    mcp.WithObject("stack", mcp.Description("CSS stack metadata")),
), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    args := argsMap(req)
    input := tools.MiToolInput{
        Param1: mcp.ParseString(req, "param1", ""),
    }
    if stackMap, ok := args["stack"].(map[string]interface{}); ok {
        input.Stack.CSSMode = strVal(stackMap, "css_mode")
        input.Stack.Framework = strVal(stackMap, "framework")
    }
    return mcp.NewToolResultText(app.srv.MiTool(ctx, input)), nil
})
```

**Paso 3 — Añadir tests**

Crear `internal/tools/mi_tool_test.go` con un `TestMiTool` que use `testutil.NewTestDB(t)` para la DB y stubs para dpf/embedder si aplica.

**Paso 4 — Documentar**

Añadir la entrada a la tabla de tools en `AGENTS.md` y crear (opcionalmente) un fichero en `docs/`.

### 6.2 Añadir soporte para un nuevo framework

En `internal/tools/suggest_layout.go`, añadir un `case "mi-framework":` en el switch de `generateLayout()` con una función `generateMiFrameworkLayout(name, desc, cssMode, fidelity)`.

Los tests en `suggest_layout_test.go` cubren los frameworks existentes; añadir un caso de prueba equivalente.

### 6.3 Añadir nuevas paletas de color

En `internal/tools/suggest_color_palettes.go`, añadir entradas al slice `predefinedPalettes` con su `paletteTemplate` (name, description, mood, useCaseTags, tokens). El algoritmo de scoring las incluirá automáticamente.

### 6.4 Añadir una nueva vista a la TUI

1. Añadir una constante `ViewMiVista` en el tipo `View` en `internal/tui/model.go`.
2. Crear `internal/tui/mi_vista.go` con un struct `miVistaModel` que implemente `Init()`, `Update()`, `View()` y un campo `goHome bool`.
3. Añadir el sub-model al `Model` raíz en `model.go`.
4. Inicializarlo en `New()`.
5. Añadir el case en `Update()` y `View()` del modelo raíz.
6. Añadir el case en `homeItemToView()` si es navegable desde el menú principal.

### 6.5 Añadir una nueva tabla a la base de datos

Añadir el `CREATE TABLE IF NOT EXISTS` en `internal/db/schema.go` (o el fichero de migraciones correspondiente). La función `db.Open()` ya ejecuta las migraciones al arrancar — es idempotente.

Si la tabla requiere búsqueda: añadir también la tabla virtual FTS5 (`CREATE VIRTUAL TABLE ... USING fts5(...)`) y el índice vector si aplica.

### 6.6 Añadir un nuevo tipo de job al bridge Rust

1. Crear un nuevo struct en `internal/dpf/jobs.go` que implemente la interfaz `Job`.
2. El binario Rust debe soportar el nuevo `operation` en su dispatcher.
3. Crear el handler de tool en `internal/tools/` que use `s.Imgproc.Execute(job)`.

### 6.7 Puntos de extensión clave

| Punto | Fichero | Qué añadir |
|-------|---------|------------|
| Nuevo tool MCP | `internal/tools/` + `cmd/devforge-mcp/main.go` | Handler + registro |
| Nuevo framework | `internal/tools/suggest_layout.go` | Case en switch |
| Nueva paleta | `internal/tools/suggest_color_palettes.go` | Entry en `predefinedPalettes` |
| Nueva vista TUI | `internal/tui/` + `model.go` | Sub-model + navegación |
| Nueva tabla DB | `internal/db/schema.go` | CREATE TABLE IF NOT EXISTS |
| Nuevo job Rust | `internal/dpf/jobs.go` | Struct + Rust handler |
| Config nueva opción | `internal/config/config.go` | Campo en `Config` struct |

---

## Referencia rápida de build y operación

```bash
# Build todo
CGO_ENABLED=1 go build ./...

# Build + instalar en ~/.local/bin/
make install

# Build + DB init + seed
make dist

# Tests
CGO_ENABLED=1 go test ./...

# Smoke test del servidor MCP
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  | ./dist/devforge-mcp

# Ejecutar la TUI
make tui
```

### Configuración mínima

```bash
mkdir -p ~/.config/devforge
cat > ~/.config/devforge/config.json <<'EOF'
{
  "gemini_api_key": "",
  "ollama_url": "http://localhost:11434",
  "embedding_model": "nomic-embed-text"
}
EOF
chmod 600 ~/.config/devforge/config.json
```

### Conectar a Claude Code

```bash
claude mcp add devforge ~/.local/bin/devforge-mcp \
  -e DEV_FORGE_CONFIG=~/.config/devforge/config.json
```

---

*Para referencia detallada de cada tool ver `docs/mcp-server-devforge.md`. Para instrucciones de conexión con clientes MCP ver `docs/mcp-connect.md`. Para la guía de instalación completa ver `docs/install.md`.*
