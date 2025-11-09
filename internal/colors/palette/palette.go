package palette

import (
	"fmt"
	"sort"

	"github.com/0mega24/go-wall/internal/colors"
)

type ByBrightness []colors.Centroid

func (b ByBrightness) Len() int { return len(b) }

func (b ByBrightness) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b ByBrightness) Less(i, j int) bool {
	brightness := func(c colors.Centroid) float32 {
		return 0.299*float32(c.R) + 0.587*float32(c.G) + 0.114*float32(c.B)
	}
	return brightness(b[i]) < brightness(b[j])
}

func FilterSimilar(colorsList []colors.Centroid, threshold float32) []colors.Centroid {
	if len(colorsList) == 0 {
		return nil
	}

	filtered := []colors.Centroid{colorsList[0]}
	for _, c := range colorsList[1:] {
		tooClose := false
		for _, f := range filtered {
			if colors.DistanceSq(c, f) < threshold {
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

func GenerateTones(pixels []colors.Centroid, n int) []colors.Centroid {
	if n <= 0 || len(pixels) == 0 {
		return nil
	}

	var sumR, sumG, sumB int
	for _, c := range pixels {
		sumR += int(c.R)
		sumG += int(c.G)
		sumB += int(c.B)
	}
	avgR := float32(sumR) / float32(len(pixels))
	avgG := float32(sumG) / float32(len(pixels))
	avgB := float32(sumB) / float32(len(pixels))

	tones := make([]colors.Centroid, n)
	for i := 0; i < n; i++ {
		factor := float32(i) / float32(n-1)
		r := avgR*0.3 + avgR*0.7*factor
		g := avgG*0.3 + avgG*0.7*factor
		b := avgB*0.3 + avgB*0.7*factor
		tones[i] = colors.Centroid{
			R: uint8(clamp(r, 0, 255)),
			G: uint8(clamp(g, 0, 255)),
			B: uint8(clamp(b, 0, 255)),
		}
	}

	return tones
}

func GenerateANSI(filtered []colors.Centroid, brightness float32) []colors.Centroid {
	if len(filtered) == 0 {
		return nil
	}

	SortByBrightness(filtered)

	baseCount := 8
	if len(filtered) < baseCount {
		baseCount = len(filtered)
	}
	baseColors := make([]colors.Centroid, baseCount)
	step := float64(len(filtered)-1) / float64(baseCount-1)
	for i := 0; i < baseCount; i++ {
		idx := int(step*float64(i) + 0.5)
		baseColors[i] = filtered[idx]
	}

	ansi := make([]colors.Centroid, 16)
	copy(ansi[:baseCount], baseColors)
	for i := 0; i < baseCount; i++ {
		ansi[i+baseCount] = brighten(baseColors[i], float32(brightness))
	}

	return ansi
}

func brighten(c colors.Centroid, factor float32) colors.Centroid {
	r := float32(c.R) + (255-float32(c.R))*factor
	g := float32(c.G) + (255-float32(c.G))*factor
	b := float32(c.B) + (255-float32(c.B))*factor
	return colors.Centroid{
		R: uint8(clamp(r, 0, 255)),
		G: uint8(clamp(g, 0, 255)),
		B: uint8(clamp(b, 0, 255)),
	}
}

func clamp(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func PrintHex(colorsList []colors.Centroid) {
	for _, c := range colorsList {
		fmt.Println(c.Hex())
	}
}

func SortByBrightness(colorsList []colors.Centroid) {
	sort.Sort(ByBrightness(colorsList))
}
