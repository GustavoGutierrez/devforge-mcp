# Generating UI Images with `generate_ui_image`

The `generate_ui_image` tool generates high-fidelity UI mockups and wireframes
via Gemini Vision. You can invoke it from any MCP-connected chat (e.g. GitHub
Copilot, Claude, Cursor) by tagging `@devforge`.

**Prerequisite:** a valid Gemini API key must be configured:

```
@devforge configure_gemini with api_key: "AIza..."
```

---

## Parameters

| Parameter     | Type     | Required | Description                                        |
|---------------|----------|----------|----------------------------------------------------|
| `prompt`      | `string` | Yes      | Visual description of the UI to generate.          |
| `style`       | `string` | Yes      | `mockup` · `wireframe` · `component` · `prototype` |
| `width`       | `int`    | No       | Output width in pixels (default: 1280).            |
| `height`      | `int`    | No       | Output height in pixels (default: 800).            |
| `output_path` | `string` | Yes      | Absolute path for the PNG file.                    |

---

## Forms

### Verbose — full parameter list

Use this form when you need precise control over dimensions and output path.

```
@devforge Usa generate_ui_image con:
- prompt: "Modern analytics dashboard with a dark sidebar navigation, KPI
  summary cards (revenue, users, conversion rate, churn), a large multi-line
  chart showing monthly revenue trends in vibrant blues and purples, a colorful
  donut chart for traffic sources (Google, Direct, Social, Email) with bright
  accent colors, a bar chart comparing weekly sales by product category using a
  coral-teal-amber palette, a recent activity feed panel, and a data table with
  row striping. Clean, professional SaaS UI. Dark mode background with
  glass-morphism card effects."
- style: "mockup"
- width: 1440
- height: 900
- output_path: "/home/user/project/dashboard_demo.png"
```

### Short — natural language

Omit dimension parameters and let the tool use its defaults (1280 × 720).
`output_path` is always required.

```
@devforge generate_ui_image prompt="Login page with email/password fields,
OAuth buttons for Google and GitHub, and a split-layout hero image on the
right. Light mode, rounded corners, Inter font." style="mockup"
output_path="/home/user/project/login.png"
```

```
@devforge generate_ui_image prompt="Onboarding wizard — 3-step progress bar,
form fields, and a friendly illustration on each step." style="wireframe"
output_path="/home/user/project/onboarding.png"
```

### Compact with explicit dimensions

```
@devforge generate_ui_image prompt="E-commerce product page: hero image, price
badge, color swatches, add-to-cart CTA, and a reviews carousel."
style="mockup" width=1280 height=720
```

---

## Style reference

| Style       | Best for                                         |
|-------------|--------------------------------------------------|
| `mockup`    | High-fidelity final UI screens                   |
| `wireframe` | Low-fidelity structure, layout exploration       |
| `component` | Single isolated UI component (button, card, etc) |
| `prototype` | Multi-state interactive-looking screen           |

---

## Prompt tips

- **Describe layout zones:** sidebar, header, main content, footer.
- **Name chart types explicitly:** bar chart, donut chart, sparkline, heatmap.
- **Specify color palette:** "coral-teal-amber", "vibrant blues and purples",
  "monochrome with a single accent".
- **Set the visual theme:** dark mode, light mode, glass-morphism, neumorphism.
- **State the brand context:** "clean SaaS UI", "fintech dashboard",
  "healthcare patient portal".
- **Reference typography if needed:** "Inter font", "monospace labels",
  "large display headings".

---

## Output

The tool returns:

```json
{
  "path": "/absolute/path/to/output.png",
  "width": 1440,
  "height": 900,
  "prompt_used": "Create a high-fidelity UI mockup of: ..."
}
```

The image is saved as a PNG at the specified (or auto-generated) path.
