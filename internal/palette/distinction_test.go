package palette

import (
	"testing"

	"github.com/0mega24/gowall/internal/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetLuminance(t *testing.T) {
	c := color.Centroid{R: 128, G: 128, B: 128}
	out := SetLuminance(c, 0.5)
	assert.False(t, out.R == 0 && out.G == 0 && out.B == 0, "SetLuminance(0.5) should not produce black")
	// Full black -> target luminance
	black := color.Centroid{R: 0, G: 0, B: 0}
	out2 := SetLuminance(black, 0.2)
	assert.False(t, out2.R == 0 && out2.G == 0 && out2.B == 0, "SetLuminance(black, 0.2) should lighten")
}

func TestEnsureTonesDistinguishable(t *testing.T) {
	tones := []color.Centroid{
		{R: 40, G: 40, B: 40},
		{R: 40, G: 40, B: 40},
		{R: 200, G: 200, B: 200},
	}
	out := EnsureTonesDistinguishable(tones)
	require.Len(t, out, 3)
	for i := 1; i < len(out); i++ {
		assert.False(t, out[i].R == out[i-1].R && out[i].G == out[i-1].G && out[i].B == out[i-1].B,
			"adjacent tones %d and %d should differ", i-1, i)
	}
}

func TestEnsureDistinguishable(t *testing.T) {
	bg := color.Centroid{R: 20, G: 20, B: 20}
	colors := []color.Centroid{
		{R: 100, G: 100, B: 100},
		{R: 101, G: 101, B: 101},
	}
	out := EnsureDistinguishable(colors, MinDistSq, bg, 3.0)
	require.Len(t, out, 2)
	assert.True(t, color.DistanceSq(out[0], out[1]) >= MinDistSq, "colors should be pushed apart to meet MinDistSq")
}
