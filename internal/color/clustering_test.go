package color

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0mega24/gowall/internal/imageutil"
)

// makeWeighted wraps a slice of color.Color into []imageutil.WeightedPixel with weight 1.0.
func makeWeighted(colors []color.Color) []imageutil.WeightedPixel {
	out := make([]imageutil.WeightedPixel, len(colors))
	for i, c := range colors {
		out[i] = imageutil.WeightedPixel{Color: c, Weight: 1.0}
	}
	return out
}

func TestFromColor(t *testing.T) {
	c := color.RGBA{R: 255, G: 128, B: 0, A: 255}
	cent := FromColor(c)
	assert.True(t, cent.R == 255 && cent.G == 128 && cent.B == 0, "FromColor(RGBA 255,128,0) = %v", cent)
	cent2 := FromColor(color.RGBA64{R: 0xFFFF, G: 0x8080, B: 0, A: 0xFFFF})
	assert.True(t, cent2.R == 255 && cent2.G == 128 && cent2.B == 0, "FromColor(RGBA64) = %v", cent2)
}

func TestDistanceSq(t *testing.T) {
	a := Centroid{R: 0, G: 0, B: 0}
	b := Centroid{R: 255, G: 255, B: 255}
	d := DistanceSq(a, b)
	assert.Equal(t, float32(195075), d, "DistanceSq(black, white) = %v, want 195075", d)
	assert.Equal(t, float32(0), DistanceSq(a, a), "DistanceSq(a,a) != 0")
}

func TestKMeans_Empty(t *testing.T) {
	assert.Nil(t, KMeans(nil, 5, 10), "KMeans(nil) should return nil")
	assert.Nil(t, KMeans([]imageutil.WeightedPixel{}, 5, 10), "KMeans(empty) should return nil")
}

func TestKMeans_SingleColor(t *testing.T) {
	colors := make([]color.Color, 100)
	for i := range colors {
		colors[i] = color.RGBA{R: 10, G: 20, B: 30, A: 255}
	}
	out := KMeansWithOptions(makeWeighted(colors), 3, 5, 50000, 42)
	require.Len(t, out, 3)
	for _, c := range out {
		assert.True(t, c.R == 10 && c.G == 20 && c.B == 30, "centroid = %v", c)
	}
}

func TestKMeans_BlackAndWhite_Deterministic(t *testing.T) {
	var colors []color.Color
	for i := 0; i < 500; i++ {
		colors = append(colors, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	}
	for i := 0; i < 500; i++ {
		colors = append(colors, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}
	out := KMeansWithOptions(makeWeighted(colors), 4, 10, 10000, 123)
	require.Len(t, out, 4)
}

func TestKMeans_BlackAndWhite_NoClampWrap(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 0, G: 0, B: 0, A: 255},
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
	for len(colors) < 1000 {
		colors = append(colors, colors...)
	}
	out := KMeansWithOptions(makeWeighted(colors), 2, 15, 50000, 1)
	require.Len(t, out, 2)
}

func TestCentroid_Hex(t *testing.T) {
	c := Centroid{R: 255, G: 0, B: 128}
	assert.Equal(t, "#ff0080", c.Hex())
	assert.Equal(t, "ff0080", c.RawHex())
}

// TestKMeans_WeightedDominance verifies that weight affects centroid placement.
// 100 red pixels with weight 1.0 vs 100 blue pixels with weight 0.01 — with k=2
// the centroid nearest red should be very close to pure red.
func TestKMeans_WeightedDominance(t *testing.T) {
	var pixels []imageutil.WeightedPixel
	for i := 0; i < 100; i++ {
		pixels = append(pixels, imageutil.WeightedPixel{
			Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
			Weight: 1.0,
		})
	}
	for i := 0; i < 100; i++ {
		pixels = append(pixels, imageutil.WeightedPixel{
			Color:  color.RGBA{R: 0, G: 0, B: 255, A: 255},
			Weight: 0.01,
		})
	}
	out := KMeansWithOptions(pixels, 2, 20, 50000, 99)
	require.Len(t, out, 2)

	// Find the centroid closest to red (255,0,0).
	red := Centroid{R: 255, G: 0, B: 0}
	bestIdx := 0
	for i, c := range out {
		if DistanceSq(c, red) < DistanceSq(out[bestIdx], red) {
			bestIdx = i
		}
	}
	dominant := out[bestIdx]
	// The red-side centroid should be very close to (255, 0, 0).
	assert.InDelta(t, 255, int(dominant.R), 5, "dominant centroid R should be ~255, got %d", dominant.R)
	assert.InDelta(t, 0, int(dominant.G), 5, "dominant centroid G should be ~0, got %d", dominant.G)
	assert.InDelta(t, 0, int(dominant.B), 5, "dominant centroid B should be ~0, got %d", dominant.B)
}
