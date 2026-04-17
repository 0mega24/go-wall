package imageutil

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPixels_FullyOpaque(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	// Fill with fully opaque red.
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	pixels := Pixels(img)
	require.Len(t, pixels, 4, "fully opaque 2x2 should yield 4 pixels")
	for i, p := range pixels {
		assert.InDelta(t, 1.0, p.Weight, 0.0001, "pixel[%d] weight should be 1.0 for opaque, got %v", i, p.Weight)
	}
}

func TestPixels_FullyTransparent(t *testing.T) {
	// image.NewRGBA zero-initializes to transparent black (A=0).
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	pixels := Pixels(img)
	assert.Len(t, pixels, 0, "fully transparent image should yield 0 pixels, got %d", len(pixels))
}

func TestPixels_SemiTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	// Alpha = 128 out of 255. RGBA() returns pre-multiplied 16-bit values.
	img.Set(0, 0, color.RGBA{R: 100, G: 150, B: 200, A: 128})
	pixels := Pixels(img)
	require.Len(t, pixels, 1, "semi-transparent pixel should be included")
	assert.Less(t, pixels[0].Weight, float32(1.0), "semi-transparent weight should be < 1.0")
	assert.Greater(t, pixels[0].Weight, float32(0.0), "semi-transparent weight should be > 0.0")
}

func TestPixels_MixedAlpha(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // opaque
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 0})   // transparent — excluded
	img.Set(2, 0, color.RGBA{R: 0, G: 0, B: 255, A: 128}) // semi-transparent
	pixels := Pixels(img)
	require.Len(t, pixels, 2, "transparent pixel should be excluded; got %d", len(pixels))
}

func TestPixels_EmptyBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(10, 10, 10, 10))
	pixels := Pixels(img)
	assert.Len(t, pixels, 0, "empty image should yield 0 pixels, got %d", len(pixels))
}

func TestPixelsUnweighted(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	pixels := PixelsUnweighted(img)
	assert.Equal(t, 4, len(pixels), "PixelsUnweighted(2x2) len = %d, want 4", len(pixels))
}
