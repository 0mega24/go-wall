package palette

import (
	"image/color"
	"testing"

	col "github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/imageutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeWeightedPixels wraps []color.Color into []imageutil.WeightedPixel with weight 1.0.
func makeWeightedPixels(colors []color.Color) []imageutil.WeightedPixel {
	out := make([]imageutil.WeightedPixel, len(colors))
	for i, c := range colors {
		out[i] = imageutil.WeightedPixel{Color: c, Weight: 1.0}
	}
	return out
}

func TestFullPipeline_BlackAndWhite(t *testing.T) {
	var colors []color.Color
	for i := 0; i < 2000; i++ {
		colors = append(colors, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	}
	for i := 0; i < 2000; i++ {
		colors = append(colors, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}
	for g := 1; g < 255; g += 16 {
		for i := 0; i < 20; i++ {
			colors = append(colors, color.RGBA{R: uint8(g), G: uint8(g), B: uint8(g), A: 255})
		}
	}

	centroids := col.KMeansWithOptions(makeWeightedPixels(colors), 32, 10, 50000, 99)
	require.NotEmpty(t, centroids, "KMeans returned no centroids")
	for i, c := range centroids {
		assertInRange(t, "centroid", i, c)
	}

	filtered := FilterSimilar(centroids, 1000)
	SortByBrightness(filtered)
	for i, c := range filtered {
		assertInRange(t, "filtered", i, c)
	}

	ansiColors := GenerateANSI(filtered, 0.4)
	assert.Len(t, ansiColors, 16, "GenerateANSI len = %d", len(ansiColors))
	for i, c := range ansiColors {
		assertInRange(t, "ANSI", i, c)
	}

	tones := GenerateTones(ansiColors, 16)
	assert.Len(t, tones, 16, "GenerateTones len = %d", len(tones))
	for i, c := range tones {
		assertInRange(t, "tone", i, c)
	}
	assertInRange(t, "bg", 0, tones[0])
	assertInRange(t, "fg", 15, tones[len(tones)-1])
	t.Logf("B&W pipeline: bg=%s fg=%s", tones[0].Hex(), tones[len(tones)-1].Hex())
}

func assertInRange(t *testing.T, label string, idx int, c col.Centroid) {
	t.Helper()
	assert.True(t, int(c.R) >= 0 && int(c.R) <= 255 && int(c.G) >= 0 && int(c.G) <= 255 && int(c.B) >= 0 && int(c.B) <= 255,
		"%s[%d] = (%d,%d,%d) out of range", label, idx, c.R, c.G, c.B)
}
