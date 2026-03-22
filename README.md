# Dev Forge MCP

DevForge MCP es un servidor MCP construido en Go que actúa como el núcleo de
aceleración del ciclo de desarrollo de software. Integra un ecosistema de
herramientas, utilidades, skills y sub-agentes especializados que trabajan en
conjunto para reducir la fricción en cada fase del desarrollo — desde el diseño
de arquitectura hasta la entrega de interfaces sofisticadas y producción-ready.

Más que un generador de código, DevForge MCP es una capa de inteligencia
transversal al stack que garantiza consistencia estructural, calidad replicable
y diseños modernos y complejos a través de todos los proyectos, independientemente
de la capa tecnológica en la que se trabaje.

## Capacidades clave

- **Aceleración del flujo de trabajo** — Automatiza tareas repetitivas mediante
  herramientas especializadas y sub-agentes coordinados.
- **Consistencia estructural** — Patrones y convenciones que garantizan la misma
  arquitectura base en cada módulo, componente o servicio.
- **Diseño sofisticado y replicable** — Sistemas de diseño, paletas, estructuras
  visuales y componentes de UI listos para producción.
- **Skills y sub-agentes especializados** — Extiende sus capacidades mediante
  skills configurables y agentes autónomos para frontend, backend, arquitectura,
  documentación y QA.
- **Calidad como estándar** — Buenas prácticas y validaciones integradas en el
  propio flujo.
- **Transversal al stack** — Herramienta común para frontend, backend,
  infraestructura y automatización.

## Soporte actual de stacks frontend

Las herramientas de UI y diseño del conjunto actual cubren:

- SPA vanilla JS/TS + CSS moderno (Vite 8).
- Astro, Next.js, SvelteKit, Nuxt.js.
- Tailwind CSS v4+ con el plugin oficial de Vite:
  - Importando `@import "tailwindcss";` en un único archivo CSS.
  - Design tokens en CSS en lugar de `tailwind.config.js`.

## Componentes

- `cmd/dev-forge-mcp/`
  - Servidor MCP en Go con SQLite/FTS5, opcionalmente libSQL.
  - Tools actuales (enfoque UI/diseño):
    - `analyze_layout`
    - `suggest_layout` (Tailwind v4 o CSS moderno)
    - `manage_tokens`
    - `store_pattern`
    - `list_patterns`
    - `generate_ui_image` (requiere Gemini API key)
    - `optimize_images`
    - `generate_favicon`
    - `suggest_color_palettes`
    - `configure_gemini`

- `cmd/dev-forge/`
  - CLI/TUI en Go + Bubble Tea para:
    - Navegar patrones y arquitecturas.
    - Lanzar auditorías de layouts.
    - Generar layouts, imágenes y favicons.
    - Explorar paletas de color.
    - Configurar integraciones (Gemini API key, etc.) desde la vista Settings.

- `db/ui_patterns.db`
  - SQLite con tablas de:
    - `patterns`, `architectures`, `tokens`, `audits`, `assets`, `palettes`.
  - Tablas virtuales FTS5 para búsqueda full‑text eficiente.

- `internal/imgproc/`
  - Bridge Go hacia el motor Rust de procesamiento de imágenes.
  - Binario: `bin/devforge-imgproc`.
  - Ver [`internal/imgproc/INTEGRATION.md`](internal/imgproc/INTEGRATION.md).

## Configuración

Archivo compartido entre el MCP server y el CLI:

```
~/.config/dev-forge/config.json
```

Sobreescribible con la variable de entorno `DEV_FORGE_CONFIG`.

## Opcional

- Migrar a libSQL para vector search semántico si se necesita
  búsqueda por similitud entre descripciones de patrones, manteniendo
  la opción de ejecución 100% local con SQLite estándar.


turso dev --db-file /home/meridian/Documentos/Proyectos/Personal/dev-forge/dev-forge-mcp/dist/dev-forge.db