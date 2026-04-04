# go-wall

Generate terminal and desktop themes from your wallpaper. Extracts a palette with k-means++, applies readability and distinction heuristics, and writes configs for Alacritty, Polybar, Rofi, BSPWM, GTK, VS Code, and more.

## Quick start

```bash
go-wall -tui                              # interactive: preview, pick templates, apply
go-wall /path/to/image.png                # run on image (or current wallpaper if omitted)
go-wall -help                             # usage (also shown with no args)
```

With no image path, wallpaper is auto-detected (feh, Hyprland, swaybg, or `GO_WALL_IMAGE` / `WALLPAPER` env).

## Usage

| Flag | Description |
|------|-------------|
| `-image path` | Image to use (can appear anywhere among flags) |
| `-tui [image]` | Interactive UI: preview palette/ANSI/tones, toggle templates and retone, then apply |
| `-templates list` | Comma-separated built-in IDs (default: `alacritty,polybar,rofi`). Use `-list-templates` to see all. |
| `-template-list path` | File listing templates (one per line: `id`, or `id<TAB>output`, or `file<TAB>tmpl<TAB>out`). `#` and empty lines ignored. |
| `-retone-ansi` | Use standard 16 ANSI colors retoned to the wallpaper |
| `-ansi-colors path` | File with 16 hex colors (one per line, order 0–7 normal then 8–15 bright). Retoned to wallpaper (implies `-retone-ansi`). |
| `-list-templates` | Print built-in IDs, default paths, and template variable reference; then exit |
| `-quiet` | Minimal output (no progress or color swatches) |
| `-seed n` | Fixed RNG seed for deterministic palette |
| `-help` | Show usage and exit |

Image can be given as `-image PATH` or as a positional argument after flags (e.g. `go-wall -retone-ansi ~/wall.png`).

## Color outputs

All hex strings are without `#` (e.g. `1a1a1a`). Templates and the library receive:

| Output | Description |
|--------|-------------|
| **Background** | Single hex (darkest tone). |
| **Foreground** | Single hex (lightest tone). |
| **Ansi** | 16 colors: indices 0–7 = normal (black, red, green, yellow, blue, magenta, cyan, white), 8–15 = bright. |
| **Tones** | 16 steps dark→light from the image palette (bg/fg ramp). |
| **TonesFromANSI** | 16 steps from retoned ANSI ramp; only set when `-retone-ansi` or `-ansi-colors` is used. |

In templates: `{{ .Background }}`, `{{ index .Ansi 0 }}` … `{{ index .Ansi 15 }}`, `{{ index .Tones 0 }}` … `{{ index .Tones 15 }}`, and when retone is used `{{ index .TonesFromANSI 0 }}` … `{{ index .TonesFromANSI 15 }}`. Run `go-wall -list-templates` for the full variable reference.

## Library

```go
import "github.com/0mega24/go-wall/pkg/gowall"

path, _ := gowall.WallpaperPath()
result, err := gowall.RunFromPath(path, gowall.DefaultOptions())
// or: result, err := gowall.RunFromImage(img, gowall.DefaultOptions())

// result.Theme: .Background, .Foreground, .Ansi, .Tones, .TonesFromANSI (hex slices/strings)
// result.Filtered, .ANSI, .Tones, .TonesFromANSI: color palettes for display
```

## Project layout

- `cmd/go-wall` — CLI.
- `internal/cli` — Run, list, TUI, display.
- `internal/pipeline` — Image → k-means → filter → ANSI/tones → theme.
- `internal/wallpaper` — Wallpaper detection (feh, hyprland, swaybg, env).
- `internal/color`, `internal/palette`, `internal/themes`, `internal/imageutil` — Core logic.
- `pkg/gowall` — Public API.

## License

MIT
