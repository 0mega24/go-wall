package cli

import (
	"strings"
	"testing"

	"github.com/0mega24/gowall/internal/color"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestSwatch(t *testing.T) {
	c := color.Centroid{R: 255, G: 0, B: 0}
	s := Swatch(c)
	// lipgloss may strip ANSI in non-TTY environments; just verify it returns a string
	assert.NotEmpty(t, s, "Swatch should return non-empty string: %q", s)
}

func TestSwatchHex(t *testing.T) {
	c := color.Centroid{R: 0, G: 128, B: 255}
	s := SwatchHex(c)
	assert.True(t, strings.Contains(s, "0080ff"), "SwatchHex should contain hex: %q", s)
}

func TestRowOfSwatches(t *testing.T) {
	cs := []color.Centroid{{R: 1, G: 2, B: 3}, {R: 4, G: 5, B: 6}}
	s := RowOfSwatches(cs, 2)
	assert.NotEmpty(t, s, "RowOfSwatches should be non-empty")
	s2 := RowOfSwatches(cs, 1)
	assert.NotEmpty(t, s2, "RowOfSwatches(n=1) should be non-empty")
}

func TestRowOfSwatchHex(t *testing.T) {
	cs := []color.Centroid{{R: 0, G: 0, B: 0}, {R: 255, G: 255, B: 255}}
	s := RowOfSwatchHex(cs, 2, "")
	assert.True(t, strings.Contains(s, "000000") && strings.Contains(s, "ffffff"), "RowOfSwatchHex should contain both hex: %q", s)
}

func TestSectionBoxInnerWidth(t *testing.T) {
	// Widths below the SectionBox minimum are coerced to 10 cells wide.
	assert.Equal(t, 5, SectionBoxInnerWidth(4))
	assert.Equal(t, 65, SectionBoxInnerWidth(70))
}

func TestSwatchGridCols(t *testing.T) {
	assert.Equal(t, 1, SwatchGridCols(20))
	assert.Equal(t, 2, SwatchGridCols(66))
	assert.Equal(t, 4, SwatchGridCols(130))
}

func TestTruncateLinesToDisplayWidth(t *testing.T) {
	longLine := strings.Repeat("x", 200)
	out := truncateLinesToDisplayWidth(longLine, 20)
	assert.Equal(t, 20, len(out))
	assert.Equal(t, strings.Repeat("x", 20), out)
}

func TestLipglossBorderWidthMaxWidthBehavior(t *testing.T) {
	st0 := sectionBoxStyle.Width(40)
	got0 := ansi.StringWidth(strings.Split(st0.Render("x"), "\n")[0])
	assert.Equal(t, 42, got0, "Width without MaxWidth adds 2 border columns to measured width")

	st1 := sectionBoxStyle.Width(114).MaxWidth(114)
	got1 := ansi.StringWidth(strings.Split(st1.Render("x"), "\n")[0])
	assert.Equal(t, 114, got1, "Width+MaxWidth equal caps total line to viewport width")
}

func TestSectionBoxRenderedWidthMatchesViewport(t *testing.T) {
	// SectionBox outer width must equal requested width (viewport); lipgloss Width/MaxWidth + borders are tuned together.
	out := SectionBox("Title", "hello", 60)
	for i, ln := range strings.Split(out, "\n") {
		if ln == "" {
			continue
		}
		assert.Equal(t, 60, ansi.StringWidth(ln), "line %d", i)
	}
}
