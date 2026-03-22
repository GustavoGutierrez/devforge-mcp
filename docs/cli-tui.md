# CLI/TUI: dev-forge (Go + Bubble Tea) v2

## Archivo de configuración

El CLI lee y escribe el mismo archivo compartido con el MCP server:

```
~/.config/dev-forge/config.json
```

Sobreescribible con la variable de entorno `DEV_FORGE_CONFIG`.
Si el archivo no existe se crea automáticamente al guardar.

---

## Vistas principales

1. **Home**
   - "Browse patterns"
   - "Browse architectures"
   - "Analyze layout file"
   - "Generate layout"
   - "Generate UI images"
   - "Optimize images"
   - "Generate favicon"
   - "Explore color palettes"
   - **"Settings"**
   - "Quit"

2. **Browse patterns / architectures**
   - Filtros por:
     - framework (Astro, Next, SvelteKit, Nuxt, SPA Vite, Vanilla)
     - css_mode (Tailwind v4 / CSS moderno)
     - tags, texto (vía FTS5).

3. **Analyze layout file**
   - Pregunta por:
     - ruta de archivo
     - framework
     - css_mode
   - Llama a `analyze_layout`.

4. **Generate layout**
   - Form con:
     - description
     - framework
     - css_mode
     - fidelity
   - Llama a `suggest_layout` y permite guardar como patrón.

5. **Generate UI images**
   - Si `gemini_api_key` no está configurado, muestra un aviso:
     > ⚠️  Gemini API key no configurada. Ve a Settings para habilitarla.
     Y ofrece navegar directamente a la vista Settings.
   - Si está configurado, muestra el form:
     - prompt
     - style (wireframe / mockup / illustration)
     - dimensiones (width × height)
     - output path
   - Llama a `generate_ui_image`.

6. **Optimize images**
   - Selección de:
     - paths de entrada (múltiples).
     - formatos de salida (webp, avif, png).
     - dimensiones y calidad.
   - Llama a `optimize_images` y muestra resultados.

7. **Generate favicon**
   - Selección de:
     - imagen base.
     - color de fondo.
   - Llama a `generate_favicon` y muestra paths + snippets HTML.

8. **Explore color palettes**
   - Form:
     - use_case
     - mood
   - Llama a `suggest_color_palettes`.
   - Permite guardar paletas en la tabla `palettes`.

9. **Settings**
   - Muestra el estado actual de la configuración (`~/.config/dev-forge/config.json`).
   - **Gemini API key**
     - Muestra "✓ Configurada" o "✗ No configurada".
     - Campo de texto para ingresar o actualizar el key (input enmascarado).
     - Al confirmar: guarda en `config.json` con permisos `0600` y llama
       a `configure_gemini` en el MCP server para que lo aplique en caliente.
   - Permite borrar el key (deshabilita `generate_ui_image`).

---

## Consideraciones de stack

- El TUI puede detectar (o preguntar) el stack del proyecto actual:
  - Leer `package.json`, `astro.config.*`, `next.config.*`, `svelte.config.*`,
    `nuxt.config.*` o archivos de Vite 8.
- Usa esa información para rellenar por defecto los campos `framework` y
  `css_mode` en las llamadas al servidor MCP.
