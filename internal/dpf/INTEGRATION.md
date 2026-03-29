# Guía de integración — DevPixelForge (dpf) en un proyecto Go

Esta guía explica qué archivos debes copiar y cómo usar el motor de procesamiento
de imágenes, video y audio en Rust desde cualquier proyecto Go.

---

## 1. Qué necesitas llevarte

### Binario Rust (obligatorio)

```
dpf/target/release/dpf
```

Este es el motor que hace el procesamiento real. Debe:
- Estar compilado para la plataforma destino (`make build-rust` en Linux/macOS).
- Ser accesible por el proceso Go en tiempo de ejecución.

> Para distribuir sin dependencias del sistema usa el binario estático:
> `make build-rust-static` → `dpf/target/x86_64-unknown-linux-musl/release/dpf`

### Cliente Go (obligatorio)

El cliente Go vive en `internal/dpf/` de devforge-mcp. Contiene:
- Todos los tipos de job (`ResizeJob`, `OptimizeJob`, `VideoTranscodeJob`, `AudioNormalizeJob`, etc.)
- `Client` — cliente one-shot (un proceso por operación)
- `StreamClient` — cliente streaming (proceso Rust persistente, recomendado para servidores)
- Métodos de conveniencia: `Resize`, `Optimize`, `Convert`, `Favicon`, `Placeholder`
- **Video**: `VideoTranscode`, `VideoResize`, `VideoTrim`, `VideoThumbnail`, `VideoProfile`
- **Audio**: `AudioTranscode`, `AudioTrim`, `AudioNormalize`, `AudioSilenceTrim`

---

## 2. Cómo integrarlo en tu proyecto Go

### Paso 1 — Copiar los archivos

```bash
# En tu proyecto Go
cp /ruta/a/devpixelforge/dpf/target/release/dpf ./bin/
cp -r /ruta/a/devpixelforge/internal/dpf ./internal/dpf
```

Estructura recomendada en tu proyecto:

```
mi-proyecto/
├── bin/
│   └── dpf                     # binario Rust
├── internal/
│   └── dpf/
│       ├── dpf.go              # cliente Go principal
│       ├── audio_job.go        # tipos de jobs de audio
│       ├── video_job.go        # tipos de jobs de video
│       └── image_suite_job.go  # tipos de jobs adicionales de imagen
└── ...
```

### Paso 2 — Usar en tu código

#### Opción A: Cliente one-shot (simple, para pocas operaciones)

```go
import "mi-proyecto/internal/dpf"

client := dpf.NewClient("./bin/dpf")
client.SetTimeout(60 * time.Second)

// Resize responsivo
result, err := client.Resize(ctx, "uploads/foto.jpg", "public/img", []uint32{320, 640, 1280})

// Video transcode
result, err = client.VideoTranscode(ctx, &dpf.VideoTranscodeJob{
    Operation: "video_transcode",
    Input:     "video.mp4",
    Output:    "video.webm",
    Codec:     "vp9",
})

// Audio normalize
result, err = client.AudioNormalize(ctx, &dpf.AudioNormalizeJob{
    Operation:  "audio_normalize",
    Input:      "audio.mp3",
    Output:     "audio_normalized.mp3",
    TargetLUFS: -14.0,
})
```

#### Opción B: StreamClient (recomendado para servidores MCP o alta carga)

El `StreamClient` arranca el proceso Rust **una sola vez** y reutiliza el
canal stdin/stdout para todas las operaciones. Elimina ~5ms de overhead por
operación.

```go
import "mi-proyecto/internal/dpf"

// Inicializar una vez (p.ej. al arrancar el servidor)
sc, err := dpf.NewStreamClient("./bin/dpf")
if err != nil {
    log.Fatal(err)
}
defer sc.Close()

// Enviar trabajos concurrentemente (StreamClient es thread-safe)
result, err := sc.Execute(&dpf.ResizeJob{
    Operation: "resize",
    Input:     "uploads/foto.jpg",
    OutputDir: "public/img",
    Widths:    []uint32{320, 640, 1280},
})

result, err = sc.Execute(&dpf.VideoTranscodeJob{
    Operation: "video_transcode",
    Input:     "video.mp4",
    Output:    "video.webm",
    Codec:     "vp9",
})
```

---

## 3. Integración en un MCP server Go

Patrón recomendado para un servidor MCP:

```go
type MCPServer struct {
    dpf *dpf.StreamClient
    // ... otros campos
}

func NewMCPServer(binaryPath string) (*MCPServer, error) {
    sc, err := dpf.NewStreamClient(binaryPath)
    if err != nil {
        return nil, fmt.Errorf("failed to start dpf: %w", err)
    }
    return &MCPServer{dpf: sc}, nil
}

func (s *MCPServer) Shutdown() {
    s.dpf.Close()
}

// Handler para la tool "optimize_images"
func (s *MCPServer) handleOptimizeImages(ctx context.Context, params json.RawMessage) (any, error) {
    var req struct {
        Paths     []string `json:"paths"`
        OutputDir string   `json:"output_dir"`
    }
    if err := json.Unmarshal(params, &req); err != nil {
        return nil, err
    }

    result, err := s.dpf.Execute(&dpf.OptimizeJob{
        Operation: "optimize",
        Inputs:    req.Paths,
        OutputDir: &req.OutputDir,
        AlsoWebp:  true,
    })
    if err != nil {
        return nil, fmt.Errorf("optimization failed: %w", err)
    }

    return result, nil
}

// Handler para la tool "video_transcode"
func (s *MCPServer) handleVideoTranscode(ctx context.Context, params json.RawMessage) (any, error) {
    var req struct {
        Input   string `json:"input"`
        Output  string `json:"output"`
        Codec   string `json:"codec"`
        Bitrate string `json:"bitrate,omitempty"`
    }
    if err := json.Unmarshal(params, &req); err != nil {
        return nil, err
    }

    return s.dpf.VideoTranscode(&dpf.VideoTranscodeJob{
        Operation: "video_transcode",
        Input:     req.Input,
        Output:    req.Output,
        Codec:     req.Codec,
        Bitrate:   req.Bitrate,
    })
}

// Handler para la tool "audio_normalize"
func (s *MCPServer) handleAudioNormalize(ctx context.Context, params json.RawMessage) (any, error) {
    var req struct {
        Input      string  `json:"input"`
        Output     string  `json:"output"`
        TargetLUFS float64 `json:"target_lufs"`
    }
    if err := json.Unmarshal(params, &req); err != nil {
        return nil, err
    }

    return s.dpf.AudioNormalize(&dpf.AudioNormalizeJob{
        Operation:  "audio_normalize",
        Input:      req.Input,
        Output:     req.Output,
        TargetLUFS: req.TargetLUFS,
    })
}
```

---

## 4. Checklist de integración

- [ ] Binario `dpf` copiado y con permisos de ejecución (`chmod +x`)
- [ ] Paquete `dpf` copiado a `internal/dpf/` de tu proyecto
- [ ] Ruta al binario configurada correctamente (absoluta o relativa al CWD del proceso)
- [ ] `StreamClient` inicializado al arrancar el servidor y cerrado al apagar (`defer sc.Close()`)
- [ ] Timeout adecuado para operaciones pesadas (`client.SetTimeout(120 * time.Second)`)
- [ ] FFmpeg instalado para operaciones de video/audio

---

## 5. Resumen de tipos de job disponibles

| Tipo | Campo `operation` | Cuándo usarlo |
|------|------------------|---------------|
| **Imágenes** |||
| `ResizeJob` | `"resize"` | Generar variantes responsivas |
| `OptimizeJob` | `"optimize"` | Comprimir PNG/JPEG + generar WebP |
| `ConvertJob` | `"convert"` | Cambiar formato (SVG→PNG, PNG→WebP, etc.) |
| `FaviconJob` | `"favicon"` | Generar pack de favicons desde SVG/PNG |
| `SpriteJob` | `"sprite"` | Crear sprite sheet + CSS |
| `PlaceholderJob` | `"placeholder"` | LQIP, color dominante, gradiente CSS |
| `BatchJob` | `"batch"` | Ejecutar múltiples operaciones en paralelo |
| **Video** |||
| `VideoTranscodeJob` | `"video_transcode"` | Transcodificar video a otro códec |
| `VideoResizeJob` | `"video_resize"` | Redimensionar video |
| `VideoTrimJob` | `"video_trim"` | Recortar video por tiempo |
| `VideoThumbnailJob` | `"video_thumbnail"` | Extraer frame como imagen |
| `VideoProfileJob` | `"video_profile"` | Aplicar perfil de codificación web |
| **Audio** |||
| `AudioTranscodeJob` | `"audio_transcode"` | Convertir entre formatos de audio |
| `AudioTrimJob` | `"audio_trim"` | Recortar audio por tiempo |
| `AudioNormalizeJob` | `"audio_normalize"` | Normalizar loudness a LUFS objetivo |
| `AudioSilenceTrimJob` | `"audio_silence_trim"` | Eliminar silencio inicial/final |

---

## 6. Requisitos del Sistema

### FFmpeg (requerido para video/audio)

dpf usa FFmpeg CLI para procesamiento de video y audio. Asegúrate de tener FFmpeg instalado:

```bash
# Linux
sudo apt install ffmpeg

# macOS
brew install ffmpeg

# Verificar instalación
ffmpeg -version
```

Versión mínima recomendada: **FFmpeg 6.0+**
