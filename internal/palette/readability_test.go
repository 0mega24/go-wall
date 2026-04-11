package palette

import (
	"testing"

	"github.com/0mega24/gowall/internal/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLuminance(t *testing.T) {
	black := color.Centroid{R: 0, G: 0, B: 0}
	white := color.Centroid{R: 255, G: 255, B: 255}
	assert.True(t, Luminance(black) == 0 && Luminance(white) == 1,
		"Luminance(black)=%v Luminance(white)=%v", Luminance(black), Luminance(white))
	mid := color.Centroid{R: 128, G: 128, B: 128}
	l := Luminance(mid)
	assert.True(t, l > 0 && l < 1, "Luminance(mid) = %v", l)
}

func TestContrastRatio(t *testing.T) {
	black := color.Centroid{R: 0, G: 0, B: 0}
	white := color.Centroid{R: 255, G: 255, B: 255}
	r := ContrastRatio(black, white)
	assert.True(t, r >= 20 && r <= 22, "ContrastRatio(black, white) = %v", r)
	same := color.Centroid{R: 100, G: 100, B: 100}
	assert.Equal(t, float32(1), ContrastRatio(same, same), "ContrastRatio(same, same) != 1")
}

func TestEnsureFGBGReadable(t *testing.T) {
	bg := color.Centroid{R: 20, G: 20, B: 20}
	fg := color.Centroid{R: 30, G: 30, B: 30}
	newFG, _ := EnsureFGBGReadable(fg, bg, 4.5)
	assert.True(t, ContrastRatio(newFG, bg) >= 4.5, "contrast = %v", ContrastRatio(newFG, bg))
	white := color.Centroid{R: 255, G: 255, B: 255}
	newFG2, _ := EnsureFGBGReadable(white, bg, 4.5)
	assert.True(t, newFG2.R == 255 && newFG2.G == 255 && newFG2.B == 255, "readable fg unchanged: got %v", newFG2)
}

func TestRaiseVUntilRelativeLuminance_NoWhiteBlend(t *testing.T) {
	// Saturated red cannot reach WCAG L=1 without desaturating; V-only should cap at vivid red.
	c := color.Centroid{R: 200, G: 20, B: 30}
	out := raiseVUntilRelativeLuminance(c, 1.0)
	assert.True(t, out.R > out.G+80 && out.R > out.B+80, "expected to stay a red family color, got %#v", out)
	assert.True(t, out.G < 100 && out.B < 100, "blending to white would push G/B high: %#v", out)
}

func TestEnsureANSIReadable(t *testing.T) {
	bg := color.Centroid{R: 15, G: 15, B: 15}
	ansi := []color.Centroid{{R: 10, G: 10, B: 10}, {R: 200, G: 200, B: 200}}
	out := EnsureANSIReadable(ansi, bg, 3.0, nil)
	require.Len(t, out, 2)
	assert.True(t, ContrastRatio(out[0], bg) >= 3.0, "first ANSI contrast < 3")
	assert.Equal(t, uint8(200), out[1].R, "unchanged: got %v", out[1])
}

func TestEnsureTonesReadable(t *testing.T) {
	tones := []color.Centroid{{R: 20, G: 20, B: 20}, {R: 40, G: 40, B: 40}, {R: 50, G: 50, B: 50}}
	out := EnsureTonesReadable(tones, 4.5)
	assert.True(t, ContrastRatio(out[0], out[len(out)-1]) >= 4.5, "contrast < 4.5")
}

func TestEnsureANSIReadable_PinnedSlotsUnchanged(t *testing.T) {
	ansi := make([]color.Centroid, 16)
	for i := range ansi {
		ansi[i] = color.Centroid{R: 10, G: 10, B: 10} // very low contrast
	}
	bg := color.Centroid{R: 0, G: 0, B: 0}
	pinned := map[int]bool{1: true}
	result := EnsureANSIReadable(ansi, bg, 4.5, pinned)
	// Slot 1 should be unchanged
	assert.Equal(t, ansi[1], result[1], "pinned slot 1 should be unchanged")
	// Other slots should be adjusted (or at least not panicking)
	assert.Len(t, result, 16)
}

func TestSpreadANSILuminanceLevels_Distinct(t *testing.T) {
	bg := color.Centroid{R: 20, G: 20, B: 20}
	ansi := make([]color.Centroid, 16)
	for i := range ansi {
		ansi[i] = color.Centroid{R: 70, G: 70, B: 70}
	}
	out := SpreadANSILuminanceLevels(ansi, bg, 3.0, nil)
	require.Len(t, out, 16)
	seen := make(map[color.Centroid]bool)
	for i, c := range out {
		assert.False(t, seen[c], "duplicate at %d: %v", i, c)
		seen[c] = true
	}
}
