package palette

import (
	"math"

	"github.com/0mega24/gowall/internal/color"
)

// minVChroma is the minimum HSV value for non-achromatic colors after a global
// value scale. If value hits exactly zero, RGB becomes black and rgbToHsv loses hue
// (S becomes 0); later contrast/luminance passes only raise V, and hsvToRgb
// with S=0 yields grey — the theme looks desaturated instead of dark-vivid.
const minVChroma float32 = 0.04

func clamp01f(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ApplyGlobalHSVPercent applies global hue rotation (degrees) and multiplicative
// saturation/value scaling: S' = S * (1+satPct/100), V' = V * (1+valPct/100).
// This matches percentage semantics: -50 Val% halves luminance (V), -50 Sat%
// halves saturation — without slamming every color to black like additive -0.5 on V.
func ApplyGlobalHSVPercent(ansi []color.Centroid, hueDeg, satPct, valPct float64) []color.Centroid {
	out := make([]color.Centroid, len(ansi))
	for i := range ansi {
		r := float32(ansi[i].R) / 255
		g := float32(ansi[i].G) / 255
		b := float32(ansi[i].B) / 255
		h, s, v := rgbToHsv(r, g, b)
		if hueDeg != 0 {
			hDeg := float64(h)*360 + hueDeg
			hDeg = math.Mod(hDeg+360*10, 360)
			h = float32(hDeg / 360)
		}
		if satPct != 0 {
			f := float32(1 + satPct/100)
			s = clamp01f(s * f)
		}
		if valPct != 0 {
			f := float32(1 + valPct/100)
			v = clamp01f(v * f)
			if s > 0.02 && v < minVChroma {
				v = minVChroma
			}
		}
		rr, gg, bb := hsvToRgb(h, s, v)
		out[i] = color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
	}
	return out
}
