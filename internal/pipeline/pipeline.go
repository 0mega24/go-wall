// Package pipeline orchestrates the full gowall color extraction and theme generation workflow.
package pipeline

import (
	"fmt"
	"image"

	"github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/imageutil"
	"github.com/0mega24/gowall/internal/palette"
	"github.com/0mega24/gowall/internal/themes"
)

// GlobalAdjust applies global hue shift and multiplicative saturation/value scaling
// as percentages: S' = S×(1+SatPct/100), V' = V×(1+ValPct/100). ValPct -50 halves
// value; SatPct -50 halves saturation (relative to each color, unlike additive dV).
type GlobalAdjust struct {
	HueDeg float64 // additive rotation in degrees
	SatPct float64 // saturation scale percent (typically about -100..100)
	ValPct float64 // value scale percent (typically about -100..100)
}

// IsZero reports whether no global adjustment is requested.
func (g GlobalAdjust) IsZero() bool {
	return g.HueDeg == 0 && g.SatPct == 0 && g.ValPct == 0
}

// Options configures the color extraction and theme pipeline.
type Options struct {
	RetoneANSI      bool             // Use ANSI colors (standard or CustomANSI) retoned to palette
	CustomANSI      []color.Centroid // If RetoneANSI and len==16, use instead of StandardANSI16
	KMeansK         int
	KMeansIters     int
	MaxSamples      int
	Seed            int64 // If non-zero, use for deterministic k-means / reproducible results
	FilterThreshold float32
	Constraints     map[int]SlotConstraint // Per-slot constraints: pin, H/S/V lock, post-gen tweak
	BackgroundHex   string                 // Optional: override background color (#rrggbb)
	Clusterer       color.Clusterer        // nil = default (kmeans++)
	GlobalAdjust    GlobalAdjust           // global hue + S/V % scale; readability/spread run after (see Run)
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		KMeansK:         32,
		KMeansIters:     10,
		MaxSamples:      0,
		FilterThreshold: 1000,
	}
}

// Result holds the pipeline output: theme data and intermediate palettes for display.
type Result struct {
	Theme         themes.ThemeData
	Filtered      []color.Centroid
	ANSI          []color.Centroid
	Tones         []color.Centroid // from image palette (bg/fg ramp)
	TonesFromANSI []color.Centroid // from retoned ANSI colors (16-step ramp); only set when RetoneANSI was used
}

// ProgressFunc is called during the pipeline with a short stage name and optional detail.
type ProgressFunc func(stage, detail string)

// Run runs the full pipeline on the given image and returns the theme data and palettes.
// If progress is non-nil, it is called at each stage (e.g. "Loading", "Clustering", "Readability").
func Run(img image.Image, opts Options, progress ProgressFunc) (Result, error) {
	if progress == nil {
		progress = func(_, _ string) {}
	}
	progress("Loading", "extracting pixels")
	pixels := imageutil.Pixels(img)
	if len(pixels) == 0 {
		return Result{}, fmt.Errorf("gowall: image has no visible (non-transparent) pixels")
	}

	progress("Clustering", "k-means")
	maxSamples := opts.MaxSamples
	if maxSamples <= 0 {
		maxSamples = 50000
	}
	clusterer := opts.Clusterer
	if clusterer == nil {
		c, _ := color.Get("kmeans++")
		clusterer = c
	}
	centroids := clusterer.Cluster(pixels, opts.KMeansK, opts.KMeansIters, maxSamples, opts.Seed, nil)

	progress("Filter", "similar colors")
	filtered := palette.FilterSimilar(centroids, opts.FilterThreshold)
	palette.SortByBrightness(filtered)

	progress("Tones", "generate ramp")
	// Ramp from image palette so bg/fg match the wallpaper; compute first so retone can use bg for contrast.
	tones := palette.GenerateTones(filtered, 16)
	progress("Readability", "contrast and distinction")
	tones = palette.EnsureTonesReadable(tones, palette.MinFGBGContrast)
	tones = palette.EnsureTonesDistinguishable(tones)
	tones = palette.EnsureTonesReadable(tones, palette.MinFGBGContrast)

	// Apply background override if requested.
	if opts.BackgroundHex != "" {
		if bg, err := ParseHex(opts.BackgroundHex); err == nil {
			tones[0] = bg
		}
	}
	bgCentroid := tones[0]

	var ansiColors []color.Centroid
	if opts.RetoneANSI {
		base := palette.StandardANSI16
		if len(opts.CustomANSI) == 16 {
			base = opts.CustomANSI
			progress("ANSI", "retone custom colors")
		} else {
			progress("ANSI", "retone standard colors")
		}
		// Retone with bg so ANSI colors already meet contrast and keep the intended look.
		ansiColors = palette.RetoneANSIToPalette(base, filtered, &bgCentroid)
	} else {
		progress("ANSI", "generate from palette")
		ansiColors = palette.GenerateANSI(filtered, 0.4)
	}
	ansiColors = ApplyGlobalTweak(ansiColors, opts.GlobalAdjust)
	// Apply per-slot constraints (pins, H/S/V locks, post-gen tweaks).
	var pinnedMask map[int]bool
	ansiColors, pinnedMask = ApplyConstraints(ansiColors, opts.Constraints)

	// Harmonize after user global adjust so themes stay readable; may partially offset strong darken/brighten.
	progress("ANSI readability", "contrast vs background")
	ansiColors = palette.EnsureANSIReadable(ansiColors, bgCentroid, palette.MinANSIContrast, pinnedMask)
	progress("ANSI luminance", "spread slot levels")
	ansiColors = palette.SpreadANSILuminanceLevels(ansiColors, bgCentroid, palette.MinANSIContrast, pinnedMask)

	bg := tones[0].RawHex()
	fg := tones[len(tones)-1].RawHex()
	ansiHex := make([]string, len(ansiColors))
	for i, c := range ansiColors {
		ansiHex[i] = c.RawHex()
	}
	toneHex := make([]string, len(tones))
	for i, t := range tones {
		toneHex[i] = t.RawHex()
	}

	// Second ramp from retoned ANSI: blend toned ramp (single-hue from avg) with raw ANSI sorted by brightness.
	var tonesFromANSI []color.Centroid
	var toneFromANSIHex []string
	if opts.RetoneANSI && len(ansiColors) == 16 {
		tonedRamp := palette.GenerateTones(ansiColors, 16)
		rawRamp := make([]color.Centroid, len(ansiColors))
		copy(rawRamp, ansiColors)
		palette.SortByBrightness(rawRamp)
		tonesFromANSI = make([]color.Centroid, 16)
		for i := range tonesFromANSI {
			tonesFromANSI[i] = color.Centroid{
				R: uint8((int(tonedRamp[i].R) + int(rawRamp[i].R)) / 2),
				G: uint8((int(tonedRamp[i].G) + int(rawRamp[i].G)) / 2),
				B: uint8((int(tonedRamp[i].B) + int(rawRamp[i].B)) / 2),
			}
		}
		toneFromANSIHex = make([]string, len(tonesFromANSI))
		for i, t := range tonesFromANSI {
			toneFromANSIHex[i] = t.RawHex()
		}
	}

	theme := themes.ThemeData{
		Background:    bg,
		Foreground:    fg,
		Ansi:          ansiHex,
		Tones:         toneHex,
		TonesFromANSI: toneFromANSIHex,
	}
	return Result{
		Theme:         theme,
		Filtered:      filtered,
		ANSI:          ansiColors,
		Tones:         tones,
		TonesFromANSI: tonesFromANSI,
	}, nil
}
