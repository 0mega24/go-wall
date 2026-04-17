# Gowall TUI Help

Press **?** or **Esc** to close this overlay. Use **↑↓** to scroll.

---

## Global Keys

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch between tabs |
| `o` | Open file picker (load new wallpaper) |
| `I` | Import Gowall Color Reference file (exported palette `.txt`; skips image pipeline) |
| `?` | Toggle this help overlay |
| `q` / `Ctrl+C` | Quit |

---

## Tabs

### Config (tab 0)
Configure the pipeline before running.

| Key | Action |
|-----|--------|
| `↑` / `↓` | Move between fields |
| `←` / `→` | Cycle algorithm (on Algorithm field) |
| `-` / `+` | Decrement/increment Clusters or Iterations |
| `Space` | Toggle Retone ANSI |
| `o` | Open file picker (on Image field) |
| `Enter` / `r` | Re-run pipeline with current settings |

**Algorithms:** `kmeans++` (default), `kmeans`, `mediancut`, `octree`

**Seed:** Numeric seed for reproducible results (kmeans/kmeans++ only).

**Retone ANSI:** When enabled, ANSI slots are remapped to match the extracted palette hues.

### Adjust (tab 1)
Edit individual ANSI slot colors.

| Key | Action |
|-----|--------|
| `↑↓←→` | Navigate the 4×4 slot grid |
| `m` | Cycle slot mode: auto → lock → pin → auto |
| `Esc` | Exit pin/lock mode (reset to auto) |
| `r` | Reset slot to auto |

**Modes:**
- `auto` — pipeline assigns color automatically
- `lock` — constrain H/S/V channels: `←→` adjust hue, `[]` adjust saturation, `-+` adjust value
- `pin` — type an exact hex value; `Enter` applies it

### Templates (tab 2)
Select and preview config file templates.

| Key | Action |
|-----|--------|
| `↑` / `↓` | Move cursor through template list |
| `Space` | Toggle template selection |
| `Enter` | Apply selected templates and quit |
| `↑↓` (right pane) | Scroll template preview |

Hex values in the preview are highlighted with their actual colors.

### Palette (tab 3)
View all extracted colors with their template variable names.

| Key | Action |
|-----|--------|
| `s` | Toggle between labeled and gradient swatch mode |
| `e` | Export color reference to `./gowall-colors.txt` |

### Preview (tab 4)
Comprehensive terminal color preview showing all color groups.

| Key | Action |
|-----|--------|
| `↑↓` / `PgUp` / `PgDn` | Scroll preview |

---

## Template Variables

Use these in custom template files (`*.gowall.tmpl`):

| Variable | Description |
|----------|-------------|
| `{{ .Background }}` | Background hex (no `#`) |
| `{{ .Foreground }}` | Foreground hex (no `#`) |
| `{{ index .Ansi 0 }}` – `{{ index .Ansi 15 }}` | 16 ANSI slot colors |
| `{{ index .Tones 0 }}` – `{{ index .Tones 15 }}` | 16 tone steps (dark→light) |
| `{{ index .TonesFromANSI 0 }}` – `{{ index .TonesFromANSI 15 }}` | Retoned tone steps (when retone enabled) |

All hex values are **without** the `#` prefix. Add it in your template:
```
color = "#{{ index .Ansi 4 }}"
```

---

## Constraint Flags (CLI)

| Flag | Format | Example |
|------|--------|---------|
| `-pin-slot` | `slot=hex` | `-pin-slot 1=#cc0000` |
| `-lock-hue` | `slot=degrees` | `-lock-hue 1=0` |
| `-lock-sat` | `slot=value` | `-lock-sat 1=0.8` |
| `-lock-val` | `slot=value` | `-lock-val 0=0.05` |
| `-background` | hex | `-background #0a0a0a` |
| `-algorithm` | name | `-algorithm mediancut` |
| `-seed` | int | `-seed 12345` |

---

## Config File

`$HOME/.config/gowall/config.toml`

```toml
[defaults]
templates   = ["alacritty", "polybar"]
retone_ansi = false

[defaults.pipeline]
algorithm = "kmeans++"
k         = 32
iters     = 10

[presets.dark]
retone_ansi = true
background  = "#080808"
```

Run with preset: `gowall -preset dark wallpaper.png`
