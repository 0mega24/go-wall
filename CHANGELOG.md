# Changelog

All notable changes to this project will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- `run` command: full pipeline from image to theme templates with clustering, retone, slot constraints, and global HSV adjust
- `tui` command: interactive terminal UI with palette preview, ANSI preview, template selection, and mouse support
- `palette show` command: print current wallpaper palette as labeled swatches or gradient
- `template list/apply` commands: discover and apply built-in and user templates
- `wallpaper detect` command: print the current wallpaper path and source
- `compare` command: run multiple clustering algorithms side-by-side and save quantized images
- `steps` command: export a quantized image at each step of the clustering algorithm
- `preset list/save/apply/delete` commands: save and reuse named run configurations
- `--preset` global flag to apply a saved preset to any `run` invocation
- Clustering algorithms: kmeans++, kmeans, median cut, octree
- Wallpaper detection: feh, Hyprland, swaybg, GNOME, nitrogen, `GOWALL_IMAGE` / `WALLPAPER` env
- Built-in templates: Alacritty, Polybar, Rofi, BSPWM, GTK, VS Code
- Public API at `pkg/gowall`: `RunFromImage`, `RunFromPath`, `WallpaperPath`
- CI pipeline with golangci-lint, gofumpt, go vet, and race-enabled tests

[Unreleased]: https://github.com/0mega24/gowall/compare/main...HEAD
