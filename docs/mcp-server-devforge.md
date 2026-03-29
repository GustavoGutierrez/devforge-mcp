# MCP Server: devforge-mcp

Servidor MCP construido en Go que actúa como núcleo de aceleración del ciclo
de desarrollo de software. Integra herramientas, skills y sub-agentes que
trabajan en conjunto para reducir la fricción en cada fase del desarrollo.

El conjunto actual de tools cubre el dominio de UI y diseño (layouts, tokens,
imágenes, paletas) para proyectos con Tailwind CSS v4+ y/o CSS moderno.
La arquitectura está diseñada para ser extendida a otros dominios del stack.
## Archivo de configuración compartido

Tanto el MCP server como el CLI/TUI leen y escriben un único archivo de
configuración local:

```
~/.config/devforge/config.json
```

Estructura:

```json
{
  "gemini_api_key": "AIza..."
}
```

- La ruta puede sobreescribirse con la variable de entorno `DEV_FORGE_CONFIG`.
- Si el archivo no existe, se crea al guardar la primera clave.
- El archivo sólo debe ser legible por el usuario (`chmod 600`).

---
## Metadatos de stack

Muchos tools aceptan el objeto `stack`:

```json
{
  "css_mode": "tailwind-v4|plain-css",
  "framework": "spa-vite|astro|next|sveltekit|nuxt|vanilla"
}
```

## Tool: analyze_layout

Analiza un layout (Tailwind v4 o CSS moderno) y devuelve un informe estructurado.

**Input (JSON):**

```json
{
  "markup": "<string> marcado HTML/JSX/Svelte/Vue...",
  "stack": {
    "css_mode": "tailwind-v4",
    "framework": "next"
  },
  "page_type": "landing|dashboard|form|marketing|other",
  "device_focus": "desktop|mobile|responsive"
}
```

**Output (JSON):** (igual que antes, pero los `issues` pueden
referirse a clases Tailwind, custom properties, etc.)

---

## Tool: suggest_layout

Genera una propuesta de layout adaptada al stack.

**Input:**

```json
{
  "description": "string",
  "stack": {
    "css_mode": "plain-css",
    "framework": "astro"
  },
  "fidelity": "wireframe|mid|production",
  "tokens_profile": "string opcional"
}
```

**Output:**

```json
{
  "layout_name": "string",
  "files": [
    {
      "path": "src/pages/index.astro",
      "snippet": "string"
    }
  ],
  "css_snippets": [
    {
      "path": "src/styles/global.css",
      "snippet": "string"
    }
  ],
  "rationale": "string"
}
```

Para `css_mode = "tailwind-v4"`:

- `css_snippets` puede incluir capas de tokens según el patrón oficial
  de Tailwind v4 con Vite (importar `"tailwindcss"` en un CSS y
  definir tokens en capas CSS si se necesita).

---

## Tool: manage_tokens

Gestiona design tokens para Tailwind v4 y CSS moderno.

**Input:**

```json
{
  "mode": "read|plan-update|apply-update",
  "css_mode": "tailwind-v4|plain-css",
  "scope": "colors|spacing|typography|all",
  "proposal": {
    "colors": {
      "brand-primary": "#0ea5e9"
    },
    "spacing": {
      "space-18": "4.5rem"
    },
    "typography": {
      "font-size-base": "1rem"
    }
  }
}
```

**Output:**

```json
{
  "current_tokens": {
    "colors": {},
    "spacing": {},
    "typography": {}
  },
  "diff": "string - diff de CSS o archivos de tokens",
  "instructions": "string - pasos sugeridos para aplicar cambios por stack"
}
```

El servidor se encarga de mapear:

- Para Tailwind v4: tokens a reglas CSS compatibles con el plugin de Vite,
  sin depender de `tailwind.config.js` legacy.
- Para CSS moderno: tokens a custom properties (`:root { --color-primary: ... }`).

---

## Tool: store_pattern

(igual que antes, pero `framework` incluye astro/next/sveltekit/nuxt/spa)

---

## Tool: list_patterns

(igual que antes, pero con filtros por `css_mode` y `framework`)

---

## Tool: generate_ui_image

Genera imágenes de UI mediante **Google Gen AI Go SDK / Gemini**.

> **Requiere configuración previa del API key.**
> Si `gemini_api_key` no está presente en `~/.config/devforge/config.json`,
> la tool devuelve un error descriptivo indicando cómo configurarla:
> - Vía CLI/TUI: `devforge config` → vista **Settings**.
> - Vía MCP: llamar a `configure_gemini` con el `api_key`.

**Input:**

```json
{
  "prompt": "string",
  "style": "wireframe|mockup|illustration",
  "width": 1280,
  "height": 720,
  "output_path": "assets/generated/hero.png"
}
```

**Output:**

```json
{
  "path": "assets/generated/hero.png",
  "width": 1280,
  "height": 720,
  "prompt_used": "string"
}
```

---

## Tool: configure_gemini

Guarda el API key de Gemini en el archivo de configuración local
(`~/.config/devforge/config.json`). Usar este tool cuando el agente
necesita habilitar `generate_ui_image` y el key no está configurado.

**Input:**

```json
{
  "api_key": "AIza..."
}
```

**Output:**

```json
{
  "config_path": "/home/user/.config/devforge/config.json",
  "status": "saved"
}
```

El servidor crea el directorio si no existe y escribe el archivo con
permisos `0600`. El API key queda disponible de inmediato para
`generate_ui_image` sin reiniciar el servidor.

---

## Tool: optimize_images

Optimiza y convierte imágenes para la web (WebP, AVIF, etc.), con soporte
para múltiples imágenes en paralelo.

Implementado mediante **DevForge Image Processor** — motor Rust de alto
rendimiento invocado desde Go a través de `internal/dpf/dpf.go`
(bridge hacia el binario `bin/dpf`).
Ver [`internal/imgproc/INTEGRATION.md`](../internal/dpf/INTEGRATION.md)
para detalles de integración, tipos de job disponibles y patrones de uso
con `StreamClient`.

**Input:**

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

**Output:**

```json
{
  "results": [
    {
      "source_path": "assets/raw/hero.png",
      "outputs": [
        {
          "format": "webp",
          "path": "assets/optimized/hero-1920.webp",
          "width": 1920,
          "height": 1080,
          "approx_size_kb": 120
        },
        {
          "format": "avif",
          "path": "assets/optimized/hero-1920.avif",
          "width": 1920,
          "height": 1080,
          "approx_size_kb": 90
        }
      ]
    }
  ]
}
```

---

## Tool: generate_favicon

Genera favicons modernos a partir de una imagen base o un asset existente.

Implementado mediante **DevForge Image Processor** — motor Rust invocado
desde Go a través de `internal/dpf/dpf.go` (bridge hacia el binario
`bin/dpf`), usando el job `FaviconJob`.
Ver [`internal/imgproc/INTEGRATION.md`](../internal/dpf/INTEGRATION.md)
para detalles de integración y uso con `StreamClient`.

Requisitos:

- La imagen final debe ser cuadrada.
- No se debe deformar: si la fuente no es cuadrada, se debe recortar o
  encajar con espacio (letterboxing) en el cuadrado.

**Input:**

```json
{
  "source_path": "assets/logo.png",
  "background_color": "#ffffff",
  "sizes": [16, 32, 48, 180, 192, 512],
  "formats": ["ico", "png", "svg"]
}
```

**Output:**

```json
{
  "icons": [
    {
      "size": 32,
      "format": "png",
      "path": "public/favicon-32x32.png"
    },
    {
      "size": 16,
      "format": "png",
      "path": "public/favicon-16x16.png"
    },
    {
      "size": 180,
      "format": "png",
      "path": "public/apple-touch-icon-180x180.png"
    },
    {
      "size": 192,
      "format": "png",
      "path": "public/android-chrome-192x192.png"
    },
    {
      "size": 512,
      "format": "png",
      "path": "public/icon-512x512.png"
    },
    {
      "size": 32,
      "format": "ico",
      "path": "public/favicon.ico"
    }
  ],
  "html_snippets": [
    "<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">",
    "<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon-180x180.png">",
    "<link rel="icon" type="image/x-icon" href="/favicon.ico">"
  ]
}
```

---

## Tool: suggest_color_palettes

Sugiere paletas de color adaptadas al uso del frontend.

**Input:**

```json
{
  "use_case": "saas-dashboard|marketing-landing|ecommerce|internal-tool|blog|other",
  "brand_keywords": ["fintech", "trust", "premium"],
  "mood": "vibrant|minimal|playful|serious|dark",
  "count": 3
}
```

**Output:**

```json
{
  "palettes": [
    {
      "name": "Fintech Calm Blue",
      "description": "Paleta sobria para producto financiero de confianza",
      "tokens": {
        "background": "#0b1220",
        "surface": "#020617",
        "primary": "#22d3ee",
        "primary-soft": "#0f172a",
        "accent": "#facc15",
        "text": "#e2e8f0",
        "muted": "#475569"
      }
    }
  ]
}
```
