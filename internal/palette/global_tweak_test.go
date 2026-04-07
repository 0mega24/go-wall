package palette

import (
	"testing"

	"github.com/0mega24/gowall/internal/color"
	"github.com/stretchr/testify/assert"
)

func TestApplyGlobalHSVPercent_DarkenKeepsChroma(t *testing.T) {
	// Hue away from 0° seam (where HSV hex sector can have R=G without being grey).
	ansi := []color.Centroid{{R: 255, G: 200, B: 80}}
	out := ApplyGlobalHSVPercent(ansi, 0, 0, -90)
	// Strong relative darken must not collapse to black/grey (R=G=B).
	assert.False(t, out[0].R == out[0].G && out[0].G == out[0].B, "expected hue to survive darkening, got grey %+v", out[0])
}

func TestApplyGlobalHSVPercent_AchromaticCanGoDark(t *testing.T) {
	ansi := []color.Centroid{{R: 128, G: 128, B: 128}}
	out := ApplyGlobalHSVPercent(ansi, 0, 0, -50)
	assert.Equal(t, out[0].R, out[0].G)
	assert.Equal(t, out[0].G, out[0].B)
}
