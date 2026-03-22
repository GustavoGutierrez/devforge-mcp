package tools

import (
	"context"
	"fmt"
	"strings"
)

// SuggestLayoutInput is the input schema for the suggest_layout tool.
type SuggestLayoutInput struct {
	Description   string                 `json:"description"`
	Stack         StackMeta              `json:"stack"`
	Fidelity      string                 `json:"fidelity"` // wireframe | mid | production
	TokensProfile map[string]interface{} `json:"tokens_profile,omitempty"`
}

// FileSnippet represents a generated file with path and content.
type FileSnippet struct {
	Path    string `json:"path"`
	Snippet string `json:"snippet"`
}

// SuggestLayoutOutput is the output schema for the suggest_layout tool.
type SuggestLayoutOutput struct {
	LayoutName  string        `json:"layout_name"`
	Files       []FileSnippet `json:"files"`
	CSSSnippets []FileSnippet `json:"css_snippets"`
	Rationale   string        `json:"rationale"`
}

// SuggestLayout implements the suggest_layout MCP tool.
func (s *Server) SuggestLayout(ctx context.Context, input SuggestLayoutInput) string {
	if strings.TrimSpace(input.Description) == "" {
		return errorJSON("description is required")
	}
	if input.Stack.CSSMode == "" {
		return errorJSON("stack.css_mode is required")
	}
	if input.Stack.Framework == "" {
		return errorJSON("stack.framework is required")
	}
	if input.Fidelity == "" {
		input.Fidelity = "mid"
	}

	result := generateLayout(input)
	return mustJSON(result)
}

func generateLayout(input SuggestLayoutInput) SuggestLayoutOutput {
	desc := input.Description
	stack := input.Stack
	fidelity := input.Fidelity

	layoutName := deriveLayoutName(desc)

	var files []FileSnippet
	var cssSnippets []FileSnippet

	switch stack.Framework {
	case "astro":
		files, cssSnippets = generateAstroLayout(layoutName, desc, stack.CSSMode, fidelity)
	case "next":
		files, cssSnippets = generateNextLayout(layoutName, desc, stack.CSSMode, fidelity)
	case "sveltekit":
		files, cssSnippets = generateSvelteKitLayout(layoutName, desc, stack.CSSMode, fidelity)
	case "nuxt":
		files, cssSnippets = generateNuxtLayout(layoutName, desc, stack.CSSMode, fidelity)
	case "spa-vite":
		files, cssSnippets = generateViteLayout(layoutName, desc, stack.CSSMode, fidelity)
	default:
		files, cssSnippets = generateVanillaLayout(layoutName, desc, stack.CSSMode, fidelity)
	}

	rationale := buildRationale(input)
	return SuggestLayoutOutput{
		LayoutName:  layoutName,
		Files:       files,
		CSSSnippets: cssSnippets,
		Rationale:   rationale,
	}
}

func deriveLayoutName(desc string) string {
	words := strings.Fields(desc)
	if len(words) == 0 {
		return "CustomLayout"
	}
	name := ""
	for i, w := range words {
		if i >= 3 {
			break
		}
		name += strings.Title(strings.ToLower(w))
	}
	return name + "Layout"
}

func cssSnippetFor(cssMode, layoutName string) string {
	if cssMode == "tailwind-v4" {
		return fmt.Sprintf(`@import "tailwindcss";

@theme {
  --color-primary: #3b82f6;
  --color-background: #ffffff;
  --color-surface: #f8fafc;
  --color-text: #1e293b;
  --spacing-section: 4rem;
}

@layer base {
  body {
    @apply bg-[--color-background] text-[--color-text];
  }
}

@layer components {
  .%s-container {
    @apply max-w-7xl mx-auto px-4 sm:px-6 lg:px-8;
  }
}`, strings.ToLower(layoutName))
	}
	return fmt.Sprintf(`:root {
  --color-primary: #3b82f6;
  --color-background: #ffffff;
  --color-surface: #f8fafc;
  --color-text: #1e293b;
  --spacing-section: 4rem;
  --spacing-container: 1.5rem;
}

.%s-container {
  max-width: 80rem;
  margin: 0 auto;
  padding: 0 var(--spacing-container);
}`, strings.ToLower(layoutName))
}

func generateAstroLayout(name, desc, cssMode, fidelity string) ([]FileSnippet, []FileSnippet) {
	var markup string
	if cssMode == "tailwind-v4" {
		markup = fmt.Sprintf(`---
// %s.astro - %s
---

<section class="py-24 px-4">
  <div class="max-w-7xl mx-auto">
    <h1 class="text-5xl font-bold tracking-tight text-(--color-text)">
      Main Heading
    </h1>
    <p class="mt-6 text-xl text-(--color-muted)">
      Supporting description text goes here.
    </p>
    <div class="mt-10 flex gap-4">
      <a href="#" class="bg-(--color-primary) text-white px-8 py-3 rounded-lg font-semibold hover:opacity-90 transition">
        Get Started
      </a>
      <a href="#" class="border border-(--color-primary) text-(--color-primary) px-8 py-3 rounded-lg font-semibold hover:bg-(--color-surface) transition">
        Learn More
      </a>
    </div>
  </div>
</section>`, name, desc)
	} else {
		markup = fmt.Sprintf(`---
// %s.astro - %s
---

<section class="%s-section">
  <div class="%s-container">
    <h1 class="%s-heading">Main Heading</h1>
    <p class="%s-subtext">Supporting description text goes here.</p>
    <div class="%s-cta">
      <a href="#" class="%s-btn-primary">Get Started</a>
      <a href="#" class="%s-btn-outline">Learn More</a>
    </div>
  </div>
</section>`,
			name, desc,
			strings.ToLower(name), strings.ToLower(name), strings.ToLower(name),
			strings.ToLower(name), strings.ToLower(name), strings.ToLower(name), strings.ToLower(name))
	}

	files := []FileSnippet{{
		Path:    fmt.Sprintf("src/components/%s.astro", name),
		Snippet: markup,
	}}
	css := []FileSnippet{{
		Path:    "src/styles/global.css",
		Snippet: cssSnippetFor(cssMode, name),
	}}
	return files, css
}

func generateNextLayout(name, desc, cssMode, fidelity string) ([]FileSnippet, []FileSnippet) {
	var markup string
	if cssMode == "tailwind-v4" {
		markup = fmt.Sprintf(`// app/components/%s.tsx
// %s

export function %s() {
  return (
    <section className="py-24 px-4">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-5xl font-bold tracking-tight">
          Main Heading
        </h1>
        <p className="mt-6 text-xl text-muted-foreground">
          Supporting description text goes here.
        </p>
        <div className="mt-10 flex gap-4">
          <a href="#" className="bg-primary text-white px-8 py-3 rounded-lg font-semibold hover:opacity-90 transition">
            Get Started
          </a>
          <a href="#" className="border border-primary text-primary px-8 py-3 rounded-lg font-semibold hover:bg-surface transition">
            Learn More
          </a>
        </div>
      </div>
    </section>
  );
}`, name, desc, name)
	} else {
		markup = fmt.Sprintf(`// app/components/%s.tsx
// %s
import styles from './%s.module.css';

export function %s() {
  return (
    <section className={styles.section}>
      <div className={styles.container}>
        <h1 className={styles.heading}>Main Heading</h1>
        <p className={styles.subtext}>Supporting description text goes here.</p>
        <div className={styles.cta}>
          <a href="#" className={styles.btnPrimary}>Get Started</a>
          <a href="#" className={styles.btnOutline}>Learn More</a>
        </div>
      </div>
    </section>
  );
}`, name, desc, name, name)
	}

	files := []FileSnippet{{
		Path:    fmt.Sprintf("app/components/%s.tsx", name),
		Snippet: markup,
	}}
	css := []FileSnippet{{
		Path:    "app/globals.css",
		Snippet: cssSnippetFor(cssMode, name),
	}}
	return files, css
}

func generateSvelteKitLayout(name, desc, cssMode, fidelity string) ([]FileSnippet, []FileSnippet) {
	var markup string
	if cssMode == "tailwind-v4" {
		markup = fmt.Sprintf(`<!-- src/lib/components/%s.svelte -->
<!-- %s -->

<script lang="ts">
  export let heading = 'Main Heading';
  export let subtext = 'Supporting description text goes here.';
</script>

<section class="py-24 px-4">
  <div class="max-w-7xl mx-auto">
    <h1 class="text-5xl font-bold tracking-tight">{heading}</h1>
    <p class="mt-6 text-xl text-muted-foreground">{subtext}</p>
    <div class="mt-10 flex gap-4">
      <a href="/" class="bg-primary text-white px-8 py-3 rounded-lg font-semibold hover:opacity-90 transition">
        Get Started
      </a>
    </div>
  </div>
</section>`, name, desc)
	} else {
		markup = fmt.Sprintf(`<!-- src/lib/components/%s.svelte -->
<!-- %s -->

<section class="section">
  <div class="container">
    <h1 class="heading">Main Heading</h1>
    <p class="subtext">Supporting description text goes here.</p>
    <div class="cta">
      <a href="/" class="btn-primary">Get Started</a>
    </div>
  </div>
</section>

<style>
  .section { padding: var(--spacing-section) 1rem; }
  .container { max-width: 80rem; margin: 0 auto; }
  .heading { font-size: 3rem; font-weight: 700; color: var(--color-text); }
  .subtext { font-size: 1.25rem; color: var(--color-muted); margin-top: 1.5rem; }
  .cta { margin-top: 2.5rem; display: flex; gap: 1rem; }
  .btn-primary { background: var(--color-primary); color: white; padding: 0.75rem 2rem; border-radius: 0.5rem; text-decoration: none; }
</style>`, name, desc)
	}

	files := []FileSnippet{{
		Path:    fmt.Sprintf("src/lib/components/%s.svelte", name),
		Snippet: markup,
	}}
	css := []FileSnippet{{
		Path:    "src/app.css",
		Snippet: cssSnippetFor(cssMode, name),
	}}
	return files, css
}

func generateNuxtLayout(name, desc, cssMode, fidelity string) ([]FileSnippet, []FileSnippet) {
	var markup string
	if cssMode == "tailwind-v4" {
		markup = fmt.Sprintf(`<!-- components/%s.vue -->
<!-- %s -->

<template>
  <section class="py-24 px-4">
    <div class="max-w-7xl mx-auto">
      <h1 class="text-5xl font-bold tracking-tight">{{ heading }}</h1>
      <p class="mt-6 text-xl text-muted-foreground">{{ subtext }}</p>
      <div class="mt-10 flex gap-4">
        <NuxtLink to="/" class="bg-primary text-white px-8 py-3 rounded-lg font-semibold hover:opacity-90 transition">
          Get Started
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{ heading?: string; subtext?: string }>();
</script>`, name, desc)
	} else {
		markup = fmt.Sprintf(`<!-- components/%s.vue -->
<!-- %s -->

<template>
  <section class="section">
    <div class="container">
      <h1 class="heading">{{ heading }}</h1>
      <p class="subtext">{{ subtext }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{ heading?: string; subtext?: string }>();
</script>

<style scoped>
.section { padding: var(--spacing-section) 1rem; }
.container { max-width: 80rem; margin: 0 auto; }
.heading { font-size: 3rem; font-weight: 700; color: var(--color-text); }
.subtext { font-size: 1.25rem; color: var(--color-muted); margin-top: 1.5rem; }
</style>`, name, desc)
	}

	files := []FileSnippet{{
		Path:    fmt.Sprintf("components/%s.vue", name),
		Snippet: markup,
	}}
	css := []FileSnippet{{
		Path:    "assets/css/global.css",
		Snippet: cssSnippetFor(cssMode, name),
	}}
	return files, css
}

func generateViteLayout(name, desc, cssMode, fidelity string) ([]FileSnippet, []FileSnippet) {
	var markup string
	if cssMode == "tailwind-v4" {
		markup = fmt.Sprintf(`// src/components/%s.tsx
// %s

export function %s() {
  return (
    <section className="py-24 px-4">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-5xl font-bold tracking-tight">Main Heading</h1>
        <p className="mt-6 text-xl text-muted-foreground">Supporting text here.</p>
      </div>
    </section>
  );
}`, name, desc, name)
	} else {
		markup = fmt.Sprintf(`// src/components/%s.tsx
// %s
import './%s.css';

export function %s() {
  return (
    <section className="%s-section">
      <div className="%s-container">
        <h1 className="%s-heading">Main Heading</h1>
      </div>
    </section>
  );
}`, name, desc, strings.ToLower(name), name, strings.ToLower(name), strings.ToLower(name), strings.ToLower(name))
	}

	files := []FileSnippet{{
		Path:    fmt.Sprintf("src/components/%s.tsx", name),
		Snippet: markup,
	}}
	css := []FileSnippet{{
		Path:    "src/index.css",
		Snippet: cssSnippetFor(cssMode, name),
	}}
	return files, css
}

func generateVanillaLayout(name, desc, cssMode, fidelity string) ([]FileSnippet, []FileSnippet) {
	markup := fmt.Sprintf(`<!-- %s.html -->
<!-- %s -->

<section class="%s-section">
  <div class="%s-container">
    <h1 class="%s-heading">Main Heading</h1>
    <p class="%s-subtext">Supporting description text goes here.</p>
    <div class="%s-cta">
      <a href="#" class="%s-btn-primary">Get Started</a>
    </div>
  </div>
</section>`,
		name, desc,
		strings.ToLower(name), strings.ToLower(name), strings.ToLower(name),
		strings.ToLower(name), strings.ToLower(name), strings.ToLower(name))

	files := []FileSnippet{{
		Path:    fmt.Sprintf("%s.html", strings.ToLower(name)),
		Snippet: markup,
	}}
	css := []FileSnippet{{
		Path:    "styles.css",
		Snippet: cssSnippetFor(cssMode, name),
	}}
	return files, css
}

func buildRationale(input SuggestLayoutInput) string {
	parts := []string{
		fmt.Sprintf("Generated a %s fidelity layout for %s using %s.", input.Fidelity, input.Stack.Framework, input.Stack.CSSMode),
	}
	if input.Stack.CSSMode == "tailwind-v4" {
		parts = append(parts, "CSS tokens are defined in @theme layer — no tailwind.config.js required.")
	} else {
		parts = append(parts, "CSS custom properties defined in :root for framework-agnostic token management.")
	}
	return strings.Join(parts, " ")
}
