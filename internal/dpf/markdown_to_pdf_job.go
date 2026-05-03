package dpf

import "encoding/json"

// ThemeOverride provides typed fields for GlyphWeaveForge v0.1.6+ theme customization.
// When non-nil, the struct is serialized as the JSON value of theme_config.
type ThemeOverride struct {
	Name         string   `json:"name,omitempty"`
	BodyFontSize *float64 `json:"body_font_size_pt,omitempty"`
	CodeFontSize *float64 `json:"code_font_size_pt,omitempty"`
	HeadingScale *float64 `json:"heading_scale,omitempty"`
	MarginMM     *float64 `json:"margin_mm,omitempty"`
}

// MarkdownToPDFJob defines the Markdown-to-PDF operation contract.
type MarkdownToPDFJob struct {
	Operation      string            `json:"operation"`
	Input          string            `json:"input,omitempty"`
	MarkdownText   *string           `json:"markdown_text,omitempty"`
	MarkdownBase64 *string           `json:"markdown_base64,omitempty"`
	Output         string            `json:"output,omitempty"`
	OutputDir      string            `json:"output_dir,omitempty"`
	FileName       string            `json:"file_name,omitempty"`
	Inline         bool              `json:"inline,omitempty"`
	PageSize       *string           `json:"page_size,omitempty"`
	PageWidthMM    *float64          `json:"page_width_mm,omitempty"`
	PageHeightMM   *float64          `json:"page_height_mm,omitempty"`
	LayoutMode     *string           `json:"layout_mode,omitempty"`
	Theme          *string           `json:"theme,omitempty"`
	ThemeConfig    json.RawMessage   `json:"theme_config,omitempty"`
	// ThemeOverride provides typed fields for theme customization.
	// When set, ThemeConfig is auto-generated and used as theme_config.
	ThemeOverride *ThemeOverride    `json:"-"`
	ResourceFiles  map[string]string `json:"resource_files,omitempty"`
}

// applyThemeOverride serializes ThemeOverride into ThemeConfig before marshaling.
func (j *MarkdownToPDFJob) applyThemeOverride() {
	if j.ThemeOverride == nil {
		return
	}
	data, err := json.Marshal(j.ThemeOverride)
	if err != nil {
		return
	}
	j.ThemeConfig = json.RawMessage(data)
}
