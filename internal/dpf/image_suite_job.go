package dpf

// ─── Image Suite Jobs ─────────────────────────────────────────────────────────

// CropJob defines a crop operation on an image.
type CropJob struct {
	Operation string `json:"operation"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	X         uint32 `json:"x"`
	Y         uint32 `json:"y"`
	Width     uint32 `json:"width"`
	Height    uint32 `json:"height"`
}

// RotateJob defines a rotation/flip operation on an image.
type RotateJob struct {
	Operation string  `json:"operation"`
	Input     string  `json:"input"`
	Output    string  `json:"output"`
	Angle     float64 `json:"angle"` // degrees (90, 180, 270)
	FlipH     bool    `json:"flip_h,omitempty"`
	FlipV     bool    `json:"flip_v,omitempty"`
}

// WatermarkJob defines a watermark operation on an image.
type WatermarkJob struct {
	Operation string  `json:"operation"`
	Input     string  `json:"input"`
	Output    string  `json:"output"`
	Text      string  `json:"text,omitempty"`
	ImagePath string  `json:"image_path,omitempty"`
	Position  string  `json:"position,omitempty"` // center, tile, custom
	X         *int    `json:"x,omitempty"`
	Y         *int    `json:"y,omitempty"`
	Opacity   float64 `json:"opacity,omitempty"` // 0.0-1.0
	Size      *int    `json:"size,omitempty"`
	Color     string  `json:"color,omitempty"`
}

// AdjustJob defines image adjustments (brightness, contrast, saturation, blur, sharpen).
type AdjustJob struct {
	Operation  string  `json:"operation"`
	Input      string  `json:"input"`
	Output     string  `json:"output"`
	Brightness float64 `json:"brightness,omitempty"` // -100 to 100
	Contrast   float64 `json:"contrast,omitempty"`   // -100 to 100
	Saturation float64 `json:"saturation,omitempty"` // -100 to 100
	Blur       float64 `json:"blur,omitempty"`       // radius in pixels
	Sharpen    float64 `json:"sharpen,omitempty"`    // amount
}

// QualityJob optimizes image quality to target file size using binary search.
type QualityJob struct {
	Operation    string `json:"operation"`
	Input        string `json:"input"`
	Output       string `json:"output"`
	TargetSizeKB int    `json:"target_size_kb"`
	Format       string `json:"format,omitempty"` // webp, jpeg, png
	MaxQuality   *uint8 `json:"max_quality,omitempty"`
	MinQuality   *uint8 `json:"min_quality,omitempty"`
}

// SrcsetJob generates responsive image variants for srcset attribute.
type SrcsetJob struct {
	Operation string   `json:"operation"`
	Input     string   `json:"input"`
	OutputDir string   `json:"output_dir"`
	Widths    []uint32 `json:"widths,omitempty"`
	Sizes     []string `json:"sizes,omitempty"` // e.g., ["100vw", "(min-width: 768px) 50vw"]
	Format    string   `json:"format,omitempty"`
}

// ExifJob defines EXIF operations (strip, preserve, extract, auto_orient).
type ExifJob struct {
	Operation string `json:"operation"`
	Input     string `json:"input"`
	Output    string `json:"output,omitempty"`
	ExifOp    string `json:"exif_op"` // strip, preserve, extract, auto_orient
}
