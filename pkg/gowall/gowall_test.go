package gowall

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunFromPath(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "sample-wallpaper.png")
	if _, err := os.Stat(path); err != nil {
		t.Skip("testdata/sample-wallpaper.png not found")
	}
	result, err := RunFromPath(path, DefaultOptions())
	require.NoError(t, err)
	assert.True(t, len(result.Theme.Ansi) == 16 && len(result.Theme.Tones) == 16,
		"Theme Ansi=%d Tones=%d", len(result.Theme.Ansi), len(result.Theme.Tones))
	assert.NotEmpty(t, result.Theme.Background, "Background empty")
	assert.NotEmpty(t, result.Theme.Foreground, "Foreground empty")
}

// opaqueImage returns a small RGBA image filled with opaque pixels.
func opaqueImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 120, B: 140, A: 255})
		}
	}
	return img
}

func TestRunFromImage(t *testing.T) {
	opts := DefaultOptions()
	opts.KMeansK = 2
	result, err := RunFromImage(opaqueImage(), opts)
	require.NoError(t, err)
	assert.Len(t, result.ANSI, 16, "ANSI len = %d", len(result.ANSI))
}

func TestRunFromImage_CustomANSI(t *testing.T) {
	opts := DefaultOptions()
	opts.RetoneANSI = true
	opts.CustomANSI = nil // use standard
	result, err := RunFromImage(opaqueImage(), opts)
	require.NoError(t, err)
	assert.Len(t, result.Theme.Ansi, 16, "Theme.Ansi len = %d", len(result.Theme.Ansi))
}
