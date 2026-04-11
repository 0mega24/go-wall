package palette

import (
	"testing"

	"github.com/0mega24/gowall/internal/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateANSI_Empty(t *testing.T) {
	assert.Nil(t, GenerateANSI(nil, 0.4), "GenerateANSI(nil) should return nil")
}

func TestGenerateANSI_SingleColor_NoBlackSlots(t *testing.T) {
	filtered := []color.Centroid{{R: 80, G: 80, B: 80}}
	out := GenerateANSI(filtered, 0.4)
	require.Len(t, out, 16)
	for i, c := range out {
		assert.False(t, c.R == 0 && c.G == 0 && c.B == 0 && i > 1, "slot %d is black", i)
	}
}

func TestGenerateANSI_BlackAndWhite_AllSlotsFilled(t *testing.T) {
	filtered := []color.Centroid{{R: 0, G: 0, B: 0}, {R: 255, G: 255, B: 255}}
	out := GenerateANSI(filtered, 0.4)
	require.Len(t, out, 16)
	for i, c := range out {
		assert.True(t, int(c.R) >= 0 && int(c.R) <= 255 && int(c.G) >= 0 && int(c.G) <= 255 && int(c.B) >= 0 && int(c.B) <= 255,
			"slot %d = %v out of range", i, c)
	}
}

func TestHueSortKey_WrappedRedBeforeCyan(t *testing.T) {
	// Imperfect RGB “reds” (G≠B) get h≈0.97; must still sort before teal/cyan (~0.5).
	red := color.Centroid{R: 220, G: 25, B: 35}
	cyan := color.Centroid{R: 18, G: 204, B: 204}
	assert.Less(t, hueSortKey(red), hueSortKey(cyan), "red family should rank before cyan for ANSI ordering")
}

func TestAssignSixSlotsByHue_RankSmallestHueToSlot1(t *testing.T) {
	// Pure primaries so HSV hue order matches expectation (red has smallest h).
	red := color.Centroid{R: 255, G: 0, B: 0}
	grn := color.Centroid{R: 0, G: 255, B: 0}
	ylw := color.Centroid{R: 255, G: 255, B: 0}
	blu := color.Centroid{R: 0, G: 0, B: 255}
	mag := color.Centroid{R: 255, G: 0, B: 255}
	cya := color.Centroid{R: 0, G: 255, B: 255}
	cands := []color.Centroid{cya, blu, red, ylw, grn, mag}
	out := assignSixSlotsByHue(cands)
	assert.Equal(t, red, out[0], "smallest hue (red) maps to ANSI slot 1")
}

func TestSelectDiverseCentroids_MixHueFamilies(t *testing.T) {
	// Brightness-sorted list with a long red band then teals (similar to clustered wallpaper).
	colors := make([]color.Centroid, 0, 24)
	for i := 0; i < 12; i++ {
		colors = append(colors, color.Centroid{R: uint8(90 + i*8), G: 8, B: 14})
	}
	for i := 0; i < 12; i++ {
		colors = append(colors, color.Centroid{R: 4, G: uint8(70 + i*12), B: uint8(80 + i*12)})
	}
	SortByBrightness(colors)
	picked := selectDiverseCentroids(colors, 8)
	require.Len(t, picked, 8)
	teal := 0
	for _, c := range picked {
		if int(c.G)+int(c.B) > int(c.R)+50 {
			teal++
		}
	}
	assert.GreaterOrEqual(t, teal, 2, "expected multiple teal-family picks, got %#v", picked)
}

func TestGenerateTones_Empty(t *testing.T) {
	assert.Nil(t, GenerateTones(nil, 16), "GenerateTones(nil) should return nil")
	assert.Nil(t, GenerateTones([]color.Centroid{}, 16), "GenerateTones(empty) should return nil")
}

func TestGenerateTones_SingleTone(t *testing.T) {
	in := []color.Centroid{{R: 100, G: 100, B: 100}}
	out := GenerateTones(in, 1)
	assert.True(t, len(out) == 1 && out[0].R == 100 && out[0].G == 100 && out[0].B == 100, "GenerateTones(single) = %v", out)
}

func TestGenerateTones_BlackAndWhite_NoWrap(t *testing.T) {
	in := []color.Centroid{{R: 0, G: 0, B: 0}, {R: 255, G: 255, B: 255}}
	out := GenerateTones(in, 16)
	require.Len(t, out, 16)
	for i, c := range out {
		assert.True(t, int(c.R) >= 0 && int(c.R) <= 255 && int(c.G) >= 0 && int(c.G) <= 255 && int(c.B) >= 0 && int(c.B) <= 255,
			"tone[%d] = %v wrap", i, c)
	}
}

func TestBrighten_Clamp(t *testing.T) {
	c := color.Centroid{R: 255, G: 255, B: 255}
	out := brighten(c, 1.0)
	assert.True(t, out.R == 255 && out.G == 255 && out.B == 255, "brighten(white, 1) = %v", out)
	black := color.Centroid{R: 0, G: 0, B: 0}
	out2 := brighten(black, 0.5)
	assert.True(t, int(out2.R) >= 0 && int(out2.R) <= 255, "brighten(black, 0.5).R = %d", out2.R)
}

func TestBrighten_PreservesSaturation(t *testing.T) {
	// #9f313a — deep red; linear RGB brighten used to collapse S; V-only keeps S.
	c := color.Centroid{R: 159, G: 49, B: 58}
	r0, g0, b0 := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
	_, s0, _ := rgbToHsv(r0, g0, b0)
	out := brighten(c, 0.4)
	r1, g1, b1 := float32(out.R)/255, float32(out.G)/255, float32(out.B)/255
	_, s1, _ := rgbToHsv(r1, g1, b1)
	assert.InDelta(t, float64(s0), float64(s1), 0.02, "saturation should stay ~constant, got s0=%v s1=%v", s0, s1)
}

func TestFilterSimilar(t *testing.T) {
	in := []color.Centroid{{R: 10, G: 10, B: 10}, {R: 12, G: 12, B: 12}, {R: 200, G: 200, B: 200}}
	assert.Equal(t, 2, len(FilterSimilar(in, 100)), "FilterSimilar len should be 2")
}

func TestSortByBrightness(t *testing.T) {
	in := []color.Centroid{{R: 255, G: 255, B: 255}, {R: 0, G: 0, B: 0}, {R: 128, G: 128, B: 128}}
	SortByBrightness(in)
	assert.True(t, in[0].R == 0 && in[2].R == 255, "SortByBrightness: got %v", in)
}
