// Package gowall provides a programmatic API to extract colors from an image
// and produce theme data (background, foreground, ANSI, tones) for use in
// templates or other tooling.
package gowall

import (
	"image"

	"github.com/0mega24/gowall/internal/pipeline"
	"github.com/0mega24/gowall/internal/themes"
	"github.com/0mega24/gowall/internal/wallpaper"
)

// Options configures the extraction pipeline (k-means, readability, etc.).
type Options = pipeline.Options

// GlobalAdjust is global hue (degrees) and multiplicative Sat/Val percent (see pipeline.GlobalAdjust).
type GlobalAdjust = pipeline.GlobalAdjust

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return pipeline.DefaultOptions()
}

// Result is the output of Run: theme data and intermediate palettes.
type Result = pipeline.Result

// RunFromImage runs the full pipeline on an image and returns theme data and palettes.
func RunFromImage(img image.Image, opts Options) (Result, error) {
	return pipeline.Run(img, opts, nil)
}

// RunFromPath loads an image from path and runs the pipeline.
func RunFromPath(path string, opts Options) (Result, error) {
	img, err := wallpaper.LoadImage(path)
	if err != nil {
		return Result{}, err
	}
	return pipeline.Run(img, opts, nil)
}

// ThemeData is the struct passed to templates (Background, Foreground, Ansi, Tones).
type ThemeData = themes.ThemeData

// LoadThemeFromColorReference parses a Gowall Color Reference file (same format as the TUI palette export)
// and returns theme data for templates. Use ResultFromTheme to build a full Result for display tooling.
func LoadThemeFromColorReference(path string) (ThemeData, error) {
	return themes.LoadColorReferenceFile(path)
}

// ResultFromTheme builds a pipeline Result from imported theme data (no image; Filtered palette is empty).
func ResultFromTheme(theme ThemeData) (Result, error) {
	return pipeline.ResultFromThemeData(theme)
}

// WallpaperPath returns the current wallpaper path by trying feh, hyprland, swaybg, and env (GOWALL_IMAGE / WALLPAPER).
func WallpaperPath() (string, error) {
	return wallpaper.CurrentWallpaperPath()
}
