package palette

import (
	golorconvert "github.com/0mega24/golor/convert"
	"github.com/0mega24/gowall/internal/color"
)

const (
	MinDistSq        float32 = 1200
	MinLuminanceStep float32 = 1.12
)

// minSaturationForHue is below this we treat color as achromatic (grey) to avoid
// unstable hue from near-grey inputs (e.g. B&W images) producing a green/red tint.
const minSaturationForHue float32 = 0.02

func SetLuminance(c color.Centroid, targetL float32) color.Centroid {
	gc := centroidToGolor(c)
	hsl := golorconvert.ToHSL(gc)
	if float32(hsl.S) < minSaturationForHue {
		grey := clampByte(targetL * 255)
		return color.Centroid{R: grey, G: grey, B: grey}
	}
	hsl.L = float64(targetL)
	result := golorconvert.FromHSL(hsl)
	return color.Centroid{R: result.R8(), G: result.G8(), B: result.B8()}
}

func SpreadANSILuminanceLevels(ansi []color.Centroid, bg color.Centroid, minContrast float32, pinned map[int]bool) []color.Centroid {
	if len(ansi) != 16 {
		return ansi
	}
	lb := Luminance(bg)
	minL := minContrast*(lb+0.05) - 0.05
	if minL < 0 {
		minL = 0.01
	}
	if minL > 1 {
		minL = 1
	}
	targetL := make([]float32, 16)
	for i := 0; i < 16; i++ {
		targetL[i] = minL + (1-minL)*float32(i)/15
	}
	out := make([]color.Centroid, 16)
	for i := range ansi {
		if pinned[i] {
			out[i] = ansi[i]
			continue
		}
		c := ansi[i]
		if Luminance(c) >= targetL[i] {
			out[i] = c
		} else {
			out[i] = raiseVUntilRelativeLuminance(c, targetL[i])
		}
		if ContrastRatio(out[i], bg) < minContrast {
			out[i] = raiseVUntilContrast(out[i], bg, minContrast)
		}
	}
	return out
}

func EnsureTonesDistinguishable(tones []color.Centroid) []color.Centroid {
	if len(tones) < 2 {
		return tones
	}
	n := len(tones)
	levels := make([]float32, n)
	for i := 0; i < n; i++ {
		levels[i] = 0.12 + 0.76*float32(i)/float32(n-1)
	}
	out := make([]color.Centroid, n)
	for i := range tones {
		out[i] = SetLuminance(tones[i], levels[i])
	}
	return out
}

func EnsureDistinguishable(colorsList []color.Centroid, minDistSq float32, bg color.Centroid, minContrast float32) []color.Centroid {
	if len(colorsList) <= 1 {
		return colorsList
	}
	out := make([]color.Centroid, len(colorsList))
	copy(out, colorsList)
	white := color.Centroid{R: 255, G: 255, B: 255}
	for pass := 0; pass < 20; pass++ {
		changed := false
		for i := range out {
			for j := range out {
				if i == j || color.DistanceSq(out[i], out[j]) >= minDistSq {
					continue
				}
				for k := 0; k < 15; k++ {
					out[j] = blendToward(out[j], white, 0.08)
					if color.DistanceSq(out[i], out[j]) >= minDistSq {
						changed = true
						break
					}
				}
				if ContrastRatio(out[j], bg) < minContrast {
					out[j] = raiseVUntilContrast(out[j], bg, minContrast)
				}
			}
		}
		if !changed {
			break
		}
	}
	return out
}
