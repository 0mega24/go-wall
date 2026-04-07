package pipeline

import (
	"fmt"
	"math"

	"github.com/0mega24/golor"
	golorconvert "github.com/0mega24/golor/convert"
	"github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/palette"
)

// SlotConstraint controls how a single ANSI slot color is generated.
// All fields are optional (nil / zero = no effect).
type SlotConstraint struct {
	Pin   *color.Centroid // hard override: exact color, skips generation
	LockH *float64        // constrain hue [0,360] after generation; nil = free
	LockS *float64        // constrain saturation [0,1] after generation; nil = free
	LockV *float64        // constrain HSV value [0,1] after generation; nil = free
	Tweak SlotTweak       // post-generation delta adjustments (applied after locks)
}

// SlotTweak holds additive HSV delta adjustments applied after generation and locking.
type SlotTweak struct {
	DeltaH float64 // degrees to rotate hue (circular)
	DeltaS float64 // add to HSV saturation (clamped 0–1)
	DeltaV float64 // add to HSV value (clamped 0–1)
}

// ParseHex parses a "#rrggbb" or "rrggbb" string into a Centroid. Returns error if invalid.
func ParseHex(s string) (color.Centroid, error) {
	c, err := golor.Hex(s)
	if err != nil {
		return color.Centroid{}, fmt.Errorf("constraints: %w", err)
	}
	return color.Centroid{R: c.R8(), G: c.G8(), B: c.B8()}, nil
}

// ApplyConstraints applies all SlotConstraints to a 16-element ANSI slice.
// Pinned slots get exact colors. Locked slots have H/S/V channels overridden.
// Tweaked slots receive delta adjustments. All constrained slots are marked in
// the returned pinned set so downstream readability enforcement skips them.
func ApplyConstraints(ansi []color.Centroid, constraints map[int]SlotConstraint) ([]color.Centroid, map[int]bool) {
	if len(constraints) == 0 {
		return ansi, nil
	}
	out := make([]color.Centroid, len(ansi))
	copy(out, ansi)
	pinned := make(map[int]bool)

	for slot, sc := range constraints {
		if slot < 0 || slot >= len(out) {
			continue
		}

		// Hard pin: use exact color, no further processing.
		if sc.Pin != nil {
			out[slot] = *sc.Pin
			pinned[slot] = true
			continue
		}

		// Apply H/S/V locks and tweaks in HSV space.
		gc := golor.RGB(out[slot].R, out[slot].G, out[slot].B)
		hsv := golorconvert.ToHSV(gc)

		if sc.LockH != nil {
			hsv.H = *sc.LockH
		}
		if sc.LockS != nil {
			hsv.S = *sc.LockS
		}
		if sc.LockV != nil {
			hsv.V = *sc.LockV
		}

		// Apply tweaks.
		if sc.Tweak.DeltaH != 0 {
			hsv.H = math.Mod(hsv.H+sc.Tweak.DeltaH+360*10, 360)
		}
		if sc.Tweak.DeltaS != 0 {
			hsv.S = clamp01(hsv.S + sc.Tweak.DeltaS)
		}
		if sc.Tweak.DeltaV != 0 {
			hsv.V = clamp01(hsv.V + sc.Tweak.DeltaV)
		}

		result := golorconvert.FromHSV(hsv)
		out[slot] = color.Centroid{R: result.R8(), G: result.G8(), B: result.B8()}
		pinned[slot] = true
	}

	return out, pinned
}

// ApplyGlobalTweak applies GlobalAdjust (hue degrees + S/V percent scale) to every
// ANSI color after generation and before per-slot ApplyConstraints.
func ApplyGlobalTweak(ansi []color.Centroid, g GlobalAdjust) []color.Centroid {
	if g.IsZero() {
		return ansi
	}
	return palette.ApplyGlobalHSVPercent(ansi, g.HueDeg, g.SatPct, g.ValPct)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
