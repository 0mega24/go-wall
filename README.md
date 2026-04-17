# gowall

Generate terminal and desktop color themes from your wallpaper. Extracts a palette using k-means++, applies readability and distinction heuristics, and writes configs for Alacritty, Polybar, Rofi, BSPWM, GTK, VS Code, and more.

## Install

```sh
go install github.com/0mega24/gowall/cmd/gowall@latest
```

Or build from source:

```sh
git clone https://github.com/0mega24/gowall
cd gowall
make build        # output: bin/gowall
make install      # installs to /usr/local/bin
```

## Quick start

```sh
gowall tui                        # interactive UI: preview, pick templates, apply
gowall run                        # run on current wallpaper with default templates
gowall run ~/Pictures/wall.png    # run on a specific image
```

Wallpaper is auto-detected from feh, Hyprland, swaybg, GNOME, nitrogen, or the `GOWALL_IMAGE` / `WALLPAPER` environment variables.

## Commands

### `gowall run [image]`

Extract a palette and apply templates.

```sh
gowall run                                      # current wallpaper, default templates
gowall run ~/wall.png --templates alacritty,rofi
gowall run --retone-ansi                        # retone standard 16 ANSI to wallpaper
gowall run --algorithm mediancut                # kmeans++, kmeans, mediancut, octree
gowall run --background '#0a0a0a'              # override background color
gowall run --global-tweak '-10,0,5'            # shift hue -10°, value +5% across all slots
gowall run --preset dark                        # apply a saved preset
```

**Constraint flags** (repeatable, slot is 0-15):

| Flag | Format | Description |
|---|---|---|
| `--pin-slot` | `slot=hex` | Fix a slot to an exact color |
| `--lock-hue` | `slot=degrees` | Lock a slot's hue |
| `--lock-sat` | `slot=value` | Lock a slot's saturation |
| `--lock-val` | `slot=value` | Lock a slot's HSV value |
| `--tweak-slot` | `slot=dH,dS,dV` | Post-generation delta per slot |

**Other flags:**

| Flag | Description |
|---|---|
| `--templates` | Comma-separated template IDs (default: `alacritty,polybar,rofi`) |
| `--template-list path` | File listing templates (one per line: `id`, or `id<TAB>output`) |
| `--ansi-colors path` | File with 16 hex colors (one per line, 0-7 normal then 8-15 bright) |
| `--quiet` | Minimal output |
| `--seed n` | Fixed RNG seed for deterministic results |
| `--color-reference path` | Load theme from a Color Reference file (skips image) |

### `gowall tui [image]`

Open the interactive terminal UI. Previews the palette, ANSI colors, and tones. Lets you toggle templates, adjust retone, and apply — all without leaving the terminal.

### `gowall palette show`

Print the current wallpaper's extracted palette.

```sh
gowall palette show
gowall palette show --format gradient
```

### `gowall template list`
### `gowall template apply <id>`

Discover available templates (builtin and user-defined) and apply one directly.

```sh
gowall template list
gowall template apply alacritty
```

### `gowall wallpaper detect`

Print the detected wallpaper path and which source found it.

### `gowall preset list/save/apply/delete`

Save and reuse named run configurations.

```sh
gowall preset save dark --algorithm kmeans++ --background '#0d0d0d' --retone-ansi
gowall preset list
gowall preset apply dark ~/wall.png
gowall preset delete dark
```

Presets can also be applied inline: `gowall run --preset dark`.

### `gowall compare [image]`

Run multiple clustering algorithms side-by-side. Saves a quantized PNG and palette swatch for each algorithm.

```sh
gowall compare                                       # current wallpaper
gowall compare ~/wall.png --algorithms kmeans++,octree --k 32 --out ./out
```

### `gowall steps [image]`

Export a quantized image at every step of the chosen clustering algorithm — useful for visualizing convergence.

```sh
gowall steps --algorithm kmeans++ --k 32 --out ./steps
```

## Color outputs

All hex values are without `#` (e.g. `1a2b3c`). Templates receive:

| Variable | Description |
|---|---|
| `{{ .Background }}` | Single hex, darkest tone |
| `{{ .Foreground }}` | Single hex, lightest tone |
| `{{ index .Ansi N }}` | 16 ANSI colors (0-7 normal, 8-15 bright) |
| `{{ index .Tones N }}` | 16 tones dark to light from the palette |
| `{{ index .TonesFromANSI N }}` | 16 tones from retoned ANSI ramp (set with `--retone-ansi`) |

Run `gowall template list` for the full variable reference and template paths.

## Built-in templates

| ID | Output |
|---|---|
| `alacritty` | `~/.config/alacritty/alacritty.toml` |
| `polybar` | `~/.config/polybar/colors.ini` |
| `rofi` | `~/.config/rofi/color.rasi` |
| `bspwm` | `~/.config/bspwm/bspwmrc` |
| `gtk` | `~/.config/gtk-3.0/gtk.css` |
| `vscode` | `~/.config/Code/User/settings.json` |

## Library

```go
import "github.com/0mega24/gowall/pkg/gowall"

path, _ := gowall.WallpaperPath()
result, err := gowall.RunFromPath(path, gowall.DefaultOptions())
// or: result, err := gowall.RunFromImage(img, gowall.DefaultOptions())

// result.Theme: .Background, .Foreground, .Ansi, .Tones, .TonesFromANSI (hex strings/slices)
// result.Filtered, .ANSI, .Tones, .TonesFromANSI: color palettes for display
```

## Project layout

| Path | Description |
|---|---|
| `cmd/gowall` | CLI entry point and subcommands |
| `internal/cli` | Run, list, TUI, display, overlay |
| `internal/pipeline` | Image -> k-means -> filter -> ANSI/tones -> theme |
| `internal/color` | Clustering algorithms (kmeans++, kmeans, median cut, octree) |
| `internal/palette` | ANSI assignment, distinction, readability heuristics |
| `internal/themes` | Template discovery, install, rendering |
| `internal/wallpaper` | Wallpaper detection (feh, hyprland, swaybg, gnome, nitrogen, env) |
| `internal/config` | Config file loading, presets, merge |
| `internal/imageutil` | Pixel extraction |
| `pkg/gowall` | Public API |

## License

Apache 2.0
