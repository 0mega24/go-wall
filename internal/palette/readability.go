package palette

import (
	"github.com/0mega24/golor"
	golorcontrast "github.com/0mega24/golor/contrast"
	"github.com/0mega24/gowall/internal/color"
)

func centroidToGolor(c color.Centroid) golor.Color {
	return golor.RGB(c.R, c.G, c.B)
}

// Luminance returns the relative luminance (0–1) of c per WCAG 2.1.
func Luminance(c color.Centroid) float32 {
	return float32(golorcontrast.Luminance(centroidToGolor(c)))
}

// ContrastRatio returns the WCAG 2.1 contrast ratio between a and b.
func ContrastRatio(a, b color.Centroid) float32 {
	return float32(golorcontrast.Ratio(centroidToGolor(a), centroidToGolor(b)))
}

// Minimum WCAG contrast thresholds used throughout the palette pipeline.
const (
	MinFGBGContrast = 4.5
	MinANSIContrast = 3.0
)

// EnsureFGBGReadable lightens fg until ContrastRatio(fg, bg) >= minRatio.
func EnsureFGBGReadable(fg, bg color.Centroid, minRatio float32) (newFG, newBG color.Centroid) {
	newBG = bg
	if ContrastRatio(fg, bg) >= minRatio {
		return fg, bg
	}
	return lightenUntilContrast(fg, bg, minRatio), newBG
}

// quantizationMargin accounts for precision loss when float64 golor results are
// rounded back to uint8 centroids and re-measured in float32.
const quantizationMargin = 0.15

func lightenUntilContrast(c, bg color.Centroid, minRatio float32) color.Centroid {
	result := golorcontrast.EnforceContrast(centroidToGolor(c), centroidToGolor(bg), float64(minRatio)+quantizationMargin)
	return color.Centroid{R: result.R8(), G: result.G8(), B: result.B8()}
}

// raiseVUntilContrast increases only HSV value until contrast vs bg meets the
// threshold (same effective target as lightenUntilContrast, including quantization).
func raiseVUntilContrast(c, bg color.Centroid, minRatio float32) color.Centroid {
	threshold := minRatio + quantizationMargin
	if ContrastRatio(c, bg) >= threshold {
		return c
	}
	r, g, b := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
	h, s, v := rgbToHsv(r, g, b)
	vLow, vHigh := v, float32(1)
	var best color.Centroid
	ok := false
	for step := 0; step < 24; step++ {
		vm := (vLow + vHigh) / 2
		rr, gg, bb := hsvToRgb(h, s, vm)
		try := color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
		if ContrastRatio(try, bg) >= threshold {
			vHigh = vm
			best = try
			ok = true
		} else {
			vLow = vm
		}
	}
	if ok {
		return best
	}
	rr, gg, bb := hsvToRgb(h, s, 1)
	return color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
}

// raiseVUntilRelativeLuminance increases only HSV value until WCAG relative
// luminance meets target. If V=1 is still below target (e.g. saturated red
// cannot reach L=1), returns that fully bright saturated color instead of
// blending toward white (which was washing out ANSI slots 8–15).
func raiseVUntilRelativeLuminance(c color.Centroid, targetL float32) color.Centroid {
	if Luminance(c) >= targetL {
		return c
	}
	r, g, b := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
	h, s, v := rgbToHsv(r, g, b)
	vLow, vHigh := v, float32(1)
	var best color.Centroid
	ok := false
	for step := 0; step < 28; step++ {
		vm := (vLow + vHigh) / 2
		rr, gg, bb := hsvToRgb(h, s, vm)
		try := color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
		if Luminance(try) >= targetL {
			vHigh = vm
			best = try
			ok = true
		} else {
			vLow = vm
		}
	}
	if ok {
		return best
	}
	rr, gg, bb := hsvToRgb(h, s, 1)
	return color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
}

func blendToward(c, target color.Centroid, t float32) color.Centroid {
	r := float32(c.R) + t*(float32(target.R)-float32(c.R))
	g := float32(c.G) + t*(float32(target.G)-float32(c.G))
	b := float32(c.B) + t*(float32(target.B)-float32(c.B))
	return color.Centroid{R: clampByte(r), G: clampByte(g), B: clampByte(b)}
}

// EnsureANSIReadable raises each unpinned ANSI color until it meets minRatio contrast against bg.
func EnsureANSIReadable(ansi []color.Centroid, bg color.Centroid, minRatio float32, pinned map[int]bool) []color.Centroid {
	if len(ansi) == 0 {
		return ansi
	}
	out := make([]color.Centroid, len(ansi))
	for i, c := range ansi {
		if pinned[i] {
			out[i] = c
			continue
		}
		if ContrastRatio(c, bg) >= minRatio {
			out[i] = c
		} else {
			out[i] = raiseVUntilContrast(c, bg, minRatio)
		}
	}
	return out
}

// EnsureTonesReadable adjusts adjacent tone pairs until each meets minRatio contrast.
func EnsureTonesReadable(tones []color.Centroid, minRatio float32) []color.Centroid {
	if len(tones) < 2 || minRatio <= 0 {
		return tones
	}
	bg, fg := tones[0], tones[len(tones)-1]
	if ContrastRatio(fg, bg) >= minRatio {
		return tones
	}
	out := make([]color.Centroid, len(tones))
	copy(out, tones)
	newFG, _ := EnsureFGBGReadable(fg, bg, minRatio)
	out[len(out)-1] = newFG
	return out
}
