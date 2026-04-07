package pipeline_test

import (
	"testing"

	"github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeANSI() []color.Centroid {
	ansi := make([]color.Centroid, 16)
	for i := range ansi {
		v := uint8(i * 16)
		ansi[i] = color.Centroid{R: v, G: v, B: v}
	}
	return ansi
}

func TestApplyConstraints_Pin(t *testing.T) {
	ansi := makeANSI()
	pin := color.Centroid{R: 255, G: 0, B: 0}
	constraints := map[int]pipeline.SlotConstraint{
		1: {Pin: &pin},
	}
	out, pinned := pipeline.ApplyConstraints(ansi, constraints)
	assert.Equal(t, pin, out[1])
	assert.True(t, pinned[1])
	// Other slots unchanged
	assert.Equal(t, ansi[0], out[0])
}

func TestApplyConstraints_LockHue(t *testing.T) {
	ansi := makeANSI()
	ansi[3] = color.Centroid{R: 200, G: 100, B: 50} // a colored slot
	h := 90.0
	constraints := map[int]pipeline.SlotConstraint{
		3: {LockH: &h},
	}
	out, pinned := pipeline.ApplyConstraints(ansi, constraints)
	assert.True(t, pinned[3])
	// Result should be different from input (hue was changed)
	assert.NotEqual(t, ansi[3], out[3])
}

func TestApplyConstraints_LockSat(t *testing.T) {
	ansi := makeANSI()
	ansi[5] = color.Centroid{R: 200, G: 100, B: 50}
	s := 0.9
	constraints := map[int]pipeline.SlotConstraint{
		5: {LockS: &s},
	}
	out, pinned := pipeline.ApplyConstraints(ansi, constraints)
	assert.True(t, pinned[5])
	_ = out[5]
}

func TestApplyConstraints_Tweak(t *testing.T) {
	ansi := makeANSI()
	ansi[7] = color.Centroid{R: 200, G: 50, B: 50}
	constraints := map[int]pipeline.SlotConstraint{
		7: {Tweak: pipeline.SlotTweak{DeltaH: 30, DeltaS: 0.1, DeltaV: 0.0}},
	}
	out, pinned := pipeline.ApplyConstraints(ansi, constraints)
	assert.True(t, pinned[7])
	assert.NotEqual(t, ansi[7], out[7])
}

func TestApplyConstraints_OutOfRange(t *testing.T) {
	ansi := makeANSI()
	pin := color.Centroid{R: 255, G: 0, B: 0}
	constraints := map[int]pipeline.SlotConstraint{
		99: {Pin: &pin}, // out of range — should be ignored
	}
	out, pinned := pipeline.ApplyConstraints(ansi, constraints)
	assert.Equal(t, ansi, out)
	assert.False(t, pinned[99])
}

func TestApplyConstraints_Empty(t *testing.T) {
	ansi := makeANSI()
	out, pinned := pipeline.ApplyConstraints(ansi, nil)
	assert.Equal(t, ansi, out)
	assert.Nil(t, pinned)
}

func TestApplyGlobalTweak_Zero(t *testing.T) {
	ansi := makeANSI()
	out := pipeline.ApplyGlobalTweak(ansi, pipeline.GlobalAdjust{})
	assert.Equal(t, ansi, out)
}

func TestApplyGlobalTweak_Desaturate(t *testing.T) {
	ansi := makeANSI()
	ansi[0] = color.Centroid{R: 255, G: 0, B: 0}
	out := pipeline.ApplyGlobalTweak(ansi, pipeline.GlobalAdjust{SatPct: -50})
	assert.NotEqual(t, ansi[0], out[0])
}

func TestParseHex(t *testing.T) {
	c, err := pipeline.ParseHex("#ff0000")
	require.NoError(t, err)
	assert.Equal(t, uint8(255), c.R)
	assert.Equal(t, uint8(0), c.G)

	_, err = pipeline.ParseHex("invalid")
	assert.Error(t, err)
}
