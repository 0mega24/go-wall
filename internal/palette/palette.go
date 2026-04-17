// Package palette derives and adjusts color palettes from clustered image colors.
package palette

import (
	"fmt"
	"math"
	"sort"

	"github.com/0mega24/gowall/internal/color"
)

// ByBrightness implements sort.Interface to order centroids from darkest to brightest.
type ByBrightness []color.Centroid

func (b ByBrightness) Len() int      { return len(b) }
func (b ByBrightness) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByBrightness) Less(i, j int) bool {
	brightness := func(c color.Centroid) float32 {
		return 0.299*float32(c.R) + 0.587*float32(c.G) + 0.114*float32(c.B)
	}
	return brightness(b[i]) < brightness(b[j])
}

// FilterSimilar removes colors that are within threshold squared-distance of an already-kept color.
func FilterSimilar(colorsList []color.Centroid, threshold float32) []color.Centroid {
	if len(colorsList) == 0 {
		return nil
	}
	filtered := []color.Centroid{colorsList[0]}
	for _, c := range colorsList[1:] {
		tooClose := false
		for _, f := range filtered {
			if color.DistanceSq(c, f) < threshold {
				tooClose = true
				break
			}
		}
		if !tooClose {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// GenerateTones derives n evenly-spaced luminance tone steps from the input pixels.
func GenerateTones(pixels []color.Centroid, n int) []color.Centroid {
	if n <= 0 || len(pixels) == 0 {
		return nil
	}
	var sumR, sumG, sumB float32
	var count float32
	for _, c := range pixels {
		r := float32(c.R) / 255
		g := float32(c.G) / 255
		b := float32(c.B) / 255
		_, s, _ := rgbToHsl(r, g, b)
		weight := s
		sumR += r * weight
		sumG += g * weight
		sumB += b * weight
		count += weight
	}
	if count == 0 {
		// All inputs were achromatic (e.g. B&W image). Use equal-weight average
		// and force neutral grey so we don't amplify numerical hue noise.
		count = float32(len(pixels))
		for _, c := range pixels {
			sumR += float32(c.R) / 255
			sumG += float32(c.G) / 255
			sumB += float32(c.B) / 255
		}
	}
	avgR := sumR / count
	avgG := sumG / count
	avgB := sumB / count
	if count > 0 && count == float32(len(pixels)) {
		// Achromatic fallback: force grey so tones stay neutral.
		grey := (avgR + avgG + avgB) / 3
		avgR, avgG, avgB = grey, grey, grey
	}
	tones := make([]color.Centroid, n)
	if n == 1 {
		tones[0] = color.Centroid{R: clampByte(avgR * 255), G: clampByte(avgG * 255), B: clampByte(avgB * 255)}
		return tones
	}
	for i := 0; i < n; i++ {
		factor := float32(i) / float32(n-1)
		r := avgR * (0.2 + 0.8*factor)
		g := avgG * (0.2 + 0.8*factor)
		b := avgB * (0.2 + 0.8*factor)
		tones[i] = color.Centroid{R: clampByte(r * 255), G: clampByte(g * 255), B: clampByte(b * 255)}
	}
	return tones
}

func rgbToHsl(r, g, b float32) (h, s, l float32) {
	max := float32(math.Max(float64(r), math.Max(float64(g), float64(b))))
	min := float32(math.Min(float64(r), math.Min(float64(g), float64(b))))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2.0 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h /= 6
	return h, s, l
}

// rgbToHsv converts 0-1 RGB to HSV (h,s,v in 0-1). V = max(R,G,B), S = (max-min)/max.
func rgbToHsv(r, g, b float32) (h, s, v float32) {
	max := float32(math.Max(float64(r), math.Max(float64(g), float64(b))))
	min := float32(math.Min(float64(r), math.Min(float64(g), float64(b))))
	v = max
	if max == 0 {
		return 0, 0, 0
	}
	s = (max - min) / max
	if max == min {
		return 0, s, v
	}
	d := max - min
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h /= 6
	return h, s, v
}

// hsvToRgb converts HSV (0-1) to 0-1 RGB.
func hsvToRgb(h, s, v float32) (r, g, b float32) {
	if s <= 0 {
		return v, v, v
	}
	if h >= 1 {
		h = 0
	}
	h6 := h * 6
	i := int(h6)
	f := h6 - float32(i)
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))
	switch i % 6 {
	case 0:
		return v, t, p
	case 1:
		return q, v, p
	case 2:
		return p, v, t
	case 3:
		return p, q, v
	case 4:
		return t, p, v
	default:
		return v, p, q
	}
}

// GenerateANSI builds a 16-color ANSI palette from filtered colors at the given brightness level.
func GenerateANSI(filtered []color.Centroid, brightness float32) []color.Centroid {
	if len(filtered) == 0 {
		return nil
	}
	SortByBrightness(filtered)
	pool := append([]color.Centroid(nil), filtered...)
	base := make([]color.Centroid, 8)

	if len(pool) == 1 {
		for i := range base {
			base[i] = pool[0]
		}
	} else {
		i0 := indexMinLuminance(pool)
		i7 := indexMaxLuminance(pool)
		base[0] = pool[i0]
		base[7] = pool[i7]
		if i0 == i7 {
			for i := 1; i < 7; i++ {
				base[i] = pool[0]
			}
		} else {
			used := make([]bool, len(pool))
			used[i0] = true
			used[i7] = true
			rem := make([]color.Centroid, 0, len(pool)-2)
			for i := range pool {
				if !used[i] {
					rem = append(rem, pool[i])
				}
			}
			rem = padRemToSix(rem, pool)
			mids := assignSixSlotsByHue(rem)
			copy(base[1:7], mids[:])
		}
	}

	ansi := make([]color.Centroid, 16)
	for i := 0; i < 8; i++ {
		ansi[i] = base[i]
		ansi[i+8] = brighten(base[i], brightness)
	}
	return ansi
}

func indexMinLuminance(pool []color.Centroid) int {
	best := 0
	lb := Luminance(pool[0])
	for i := 1; i < len(pool); i++ {
		if li := Luminance(pool[i]); li < lb {
			lb = li
			best = i
		}
	}
	return best
}

func indexMaxLuminance(pool []color.Centroid) int {
	best := 0
	lb := Luminance(pool[0])
	for i := 1; i < len(pool); i++ {
		if li := Luminance(pool[i]); li > lb {
			lb = li
			best = i
		}
	}
	return best
}

func padRemToSix(rem, pool []color.Centroid) []color.Centroid {
	if len(rem) >= 6 {
		return rem
	}
	filler := pool[0]
	if len(pool) > 1 {
		filler = pool[len(pool)/2]
	}
	for len(rem) < 6 {
		rem = append(rem, filler)
	}
	return rem
}

// slotHueRankPerm[i] is which ascending-hue rank (0 = smallest in the six) fills
// ANSI slot i+1 (base index i). Canonical hue order R→Y→G→C→B→M maps ranks
// 0,1,2,3,4,5 to slots 1,3,2,6,4,5 → out[i] = sorted[slotHueRankPerm[i]].
var slotHueRankPerm = [6]int{0, 2, 1, 4, 5, 3}

func hueSortKey(c color.Centroid) float32 {
	r, g, b := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
	h, s, _ := rgbToHsv(r, g, b)
	if s < 0.04 {
		// Push greys after chromatic hues [0,1); break ties by luminance.
		return 1.0 + Luminance(c)
	}
	// Saturated reds often get h≈0.95–0.99 (near 360°). Plain numeric sort would
	// place them after cyan (~0.5), so “red” (Ansi 1) showed teal and cyans filled
	// red slots. Wrap the high red sector to negative keys so it sorts before yellow.
	if h >= 0.92 {
		return h - 1.0
	}
	return h
}

// assignSixSlotsByHue sorts 6 image colors by hue (rank order), then maps rank i
// to ANSI slots in canonical hue order (red → yellow → green → cyan → blue → magenta).
// No min-cost optimization — only reordering real centroids.
func assignSixSlotsByHue(candidates []color.Centroid) [6]color.Centroid {
	cands := append([]color.Centroid(nil), candidates...)
	if len(cands) > 14 {
		cands = selectDiverseCentroids(cands, 14)
	}
	if len(cands) < 6 {
		for len(cands) < 6 {
			cands = append(cands, cands[len(cands)-1])
		}
	}
	if len(cands) > 6 {
		cands = selectDiverseCentroids(cands, 6)
	}
	tags := make([]struct {
		c   color.Centroid
		key float32
		i   int
	}, len(cands))
	for j := range cands {
		tags[j].c = cands[j]
		tags[j].key = hueSortKey(cands[j])
		tags[j].i = j
	}
	sort.SliceStable(tags, func(a, b int) bool {
		if tags[a].key != tags[b].key {
			return tags[a].key < tags[b].key
		}
		return tags[a].i < tags[b].i
	})
	sorted := make([]color.Centroid, 6)
	for j := 0; j < 6; j++ {
		sorted[j] = tags[j].c
	}
	var out [6]color.Centroid
	for i := 0; i < 6; i++ {
		out[i] = sorted[slotHueRankPerm[i]]
	}
	return out
}

// selectDiverseCentroids picks k colors using a greedy max–min strategy (farthest-first
// in RGB). Even spacing along a brightness-sorted list often lands many indices in the
// same hue band (e.g. all reds); this spreads across distinct image colors.
func selectDiverseCentroids(colors []color.Centroid, k int) []color.Centroid {
	if k <= 0 || len(colors) == 0 {
		return nil
	}
	if len(colors) <= k {
		out := make([]color.Centroid, k)
		copy(out, colors)
		last := colors[len(colors)-1]
		for i := len(colors); i < k; i++ {
			out[i] = last
		}
		return out
	}
	used := make([]bool, len(colors))
	selected := make([]color.Centroid, 0, k)
	// Darkest anchor (colors are brightness-sorted ascending).
	selected = append(selected, colors[0])
	used[0] = true
	// Second: farthest from the anchor (often the opposite family, e.g. teal vs red).
	bestI := 1
	bestD := color.DistanceSq(colors[1], colors[0])
	for i := 2; i < len(colors); i++ {
		d := color.DistanceSq(colors[i], colors[0])
		if d > bestD {
			bestD = d
			bestI = i
		}
	}
	selected = append(selected, colors[bestI])
	used[bestI] = true
	for len(selected) < k {
		bestI = -1
		var bestMinD float32 = -1
		for i := 0; i < len(colors); i++ {
			if used[i] {
				continue
			}
			var minD float32 = 1e30
			for _, s := range selected {
				d := color.DistanceSq(colors[i], s)
				if d < minD {
					minD = d
				}
			}
			if minD > bestMinD {
				bestMinD = minD
				bestI = i
			}
		}
		if bestI < 0 {
			break
		}
		selected = append(selected, colors[bestI])
		used[bestI] = true
	}
	for len(selected) < k {
		selected = append(selected, colors[len(colors)-1])
	}
	return selected
}

// brighten raises only HSV value (V), preserving hue and saturation. The delta
// is added to V in [0,1] (same scale as GenerateANSI’s brightness argument).
// Linear RGB→white interpolation was removed because it collapses saturation.
func brighten(c color.Centroid, deltaV float32) color.Centroid {
	r, g, b := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
	h, s, v := rgbToHsv(r, g, b)
	v += deltaV
	if v > 1 {
		v = 1
	}
	if v < 0 {
		v = 0
	}
	rr, gg, bb := hsvToRgb(h, s, v)
	return color.Centroid{R: clampByte(rr * 255), G: clampByte(gg * 255), B: clampByte(bb * 255)}
}

func clampByte(val float32) uint8 {
	if val <= 0 {
		return 0
	}
	if val >= 255 {
		return 255
	}
	return uint8(val)
}

// PrintHex prints each color as a #rrggbb hex string to stdout.
func PrintHex(colorsList []color.Centroid) {
	for _, c := range colorsList {
		fmt.Println(c.Hex())
	}
}

// SortByBrightness sorts colorsList in-place from darkest to brightest.
func SortByBrightness(colorsList []color.Centroid) {
	sort.Sort(ByBrightness(colorsList))
}
