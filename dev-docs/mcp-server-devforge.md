# DevForge MCP Server

`devforge-mcp` is a **stdio-only MCP server** exposing DevForge's stateless developer tools.

Use it when an agent needs deterministic utility operations (format/convert/validate/generate/transform)
without adding app-specific business logic or external persistent state.

## What an agent should infer quickly

- **Tool selection-first server**: choose a specific tool for each narrow task instead of writing ad-hoc code.
- **Stateless by design**: no DB, no embeddings, no long-lived memory in the server.
- **Best for developer workflows**: docs/report generation, media transforms, HTTP checks, JSON/YAML/CSV conversion,
  crypto helpers, frontend/backend calculations, and date/time utilities.
- **MCP transport**: stdio only.

## Notable tool groups

- Document generation: `markdown_to_pdf`
- Media: image/video/audio tools, `generate_favicon`, `generate_ui_image`, `ui2md`
- Color utilities: `color_code_convert`, `color_harmony_palette`, `css_gradient_generate`
- Utilities: text, data, crypto, HTTP, time, file, frontend, backend, code

## High-signal tool mapping for agents

- Need **Markdown → PDF** for reports, PRPs, specs, technical docs, invoices, or deliverables:
  use `markdown_to_pdf`.
  - Input: `input` (file path) or `markdown_text` (inline) or `markdown_base64`.
  - Output: `output` (explicit path), or `output_dir` + optional `file_name`, or `inline=true` for base64 PDF.
  - Layout/theme controls: `layout_mode`, `page_size` or custom `page_width_mm/page_height_mm`, `theme`,
    plus `theme_override` (`name`, `body_font_size_pt`, `code_font_size_pt`, `heading_scale`, `margin_mm`) and
    `theme_config` for advanced override payloads.
- Need many image transforms in a single operation: `optimize_images`.
- Need screenshot to Markdown design handoff: `ui2md`.

## Runtime assumptions

- stdio only
- no database
- no Ollama or embedding services
- `dpf` must be available for media/document tools
