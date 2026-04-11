package pipeline

import (
	"image"
	"testing"

	"github.com/0mega24/gowall/internal/palette"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, image.Black)
		}
	}
	opts := DefaultOptions()
	opts.KMeansK = 2
	opts.MaxSamples = 100
	result, err := Run(img, opts, nil)
	require.NoError(t, err)
	assert.Len(t, result.Theme.Ansi, 16, "Theme.Ansi len = %d, want 16", len(result.Theme.Ansi))
	assert.Len(t, result.Theme.Tones, 16, "Theme.Tones len = %d, want 16", len(result.Theme.Tones))
	assert.NotEmpty(t, result.Theme.Background, "Background should be set")
	assert.NotEmpty(t, result.Theme.Foreground, "Foreground should be set")
	assert.NotEmpty(t, result.Filtered, "Filtered should be non-empty")
	assert.True(t, len(result.ANSI) == 16 && len(result.Tones) == 16, "ANSI=%d Tones=%d", len(result.ANSI), len(result.Tones))
}

func TestRun_RetoneStandardANSI(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for i := 0; i < 64; i++ {
		img.Set(i%8, i/8, image.White)
	}
	opts := DefaultOptions()
	opts.RetoneANSI = true
	result, err := Run(img, opts, nil)
	require.NoError(t, err)
	assert.Len(t, result.ANSI, 16, "ANSI len = %d", len(result.ANSI))
	assert.True(t, len(result.TonesFromANSI) == 16 && len(result.Theme.TonesFromANSI) == 16,
		"TonesFromANSI len = %d / %d, want 16", len(result.TonesFromANSI), len(result.Theme.TonesFromANSI))
}

func TestRun_RetoneCustomANSI(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for i := 0; i < 64; i++ {
		img.Set(i%8, i/8, image.White)
	}
	opts := DefaultOptions()
	opts.RetoneANSI = true
	opts.CustomANSI = palette.StandardANSI16 // use standard as "custom" to hit the custom path
	result, err := Run(img, opts, nil)
	require.NoError(t, err)
	assert.Len(t, result.ANSI, 16, "ANSI len = %d", len(result.ANSI))
	assert.Len(t, result.Theme.TonesFromANSI, 16, "Theme.TonesFromANSI len = %d, want 16", len(result.Theme.TonesFromANSI))
}

func TestDefaultOptions(t *testing.T) {
	o := DefaultOptions()
	assert.True(t, o.KMeansK == 32 && o.KMeansIters == 10, "DefaultOptions: K=%d Iters=%d", o.KMeansK, o.KMeansIters)
}

func TestRun_GlobalAdjustCompletes(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 12, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 12; x++ {
			img.Set(x, y, image.White)
		}
	}
	opts := DefaultOptions()
	opts.RetoneANSI = true
	opts.KMeansK = 8
	opts.MaxSamples = 2000
	opts.Seed = 7
	opts.GlobalAdjust = GlobalAdjust{ValPct: -50, SatPct: -10}
	r, err := Run(img, opts, nil)
	require.NoError(t, err)
	assert.Len(t, r.ANSI, 16)
	assert.Len(t, r.Theme.Ansi, 16)
}
