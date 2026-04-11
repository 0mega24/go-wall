package cli

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/0mega24/gowall/internal/color"
)

// Swatch renders a small color block in the terminal using lipgloss background.
func Swatch(c color.Centroid) string {
	return lipgloss.NewStyle().Background(lipgloss.Color(c.Hex())).Render("  ")
}

// SwatchHex returns a colored block followed by the hex label.
func SwatchHex(c color.Centroid) string {
	swatch := lipgloss.NewStyle().Background(lipgloss.Color(c.Hex())).Render("  ")
	return swatch + " " + c.Hex()
}

// RowOfSwatches renders up to n color swatches in a row (space-separated).
func RowOfSwatches(cs []color.Centroid, n int) string {
	if n <= 0 || n > len(cs) {
		n = len(cs)
	}
	var s string
	for i := 0; i < n; i++ {
		if i > 0 {
			s += " "
		}
		s += Swatch(cs[i])
	}
	return s
}

// RowOfSwatchHex renders swatch + hex for each color, wrapping every perLine items.
// All rows (including the first) get the indent prefix so columns align inside boxes.
func RowOfSwatchHex(cs []color.Centroid, perLine int, indent string) string {
	s := indent
	for i, c := range cs {
		if perLine > 0 && i > 0 && i%perLine == 0 {
			s += "\n" + indent
		} else if i > 0 {
			s += " "
		}
		s += SwatchHex(c)
	}
	return s
}

// GradientRow renders all colors as a continuous horizontal strip.
func GradientRow(colors []color.Centroid, width int) string {
	if len(colors) == 0 {
		return ""
	}
	perColor := width / len(colors)
	if perColor < 1 {
		perColor = 1
	}
	var sb strings.Builder
	for _, c := range colors {
		block := strings.Repeat(" ", perColor)
		sb.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(c.Hex())).Render(block))
	}
	return sb.String()
}

// sectionBoxStyle is fixed border/padding; width is applied per call in SectionBox.
// NormalBorder gives straight vertical edges; rounded corners looked like open hooks
// when paired with ANSI-heavy content.
var sectionBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#555555")).
	Padding(0, 1)

// SectionBoxInnerWidth is the number of character cells available for text inside
// a SectionBox whose width argument equals the desired inner content width.
// lipgloss Width() sets the inner content area; border (1+1) + padding (1+1) = 4
// chars of frame are added around it. We keep 1 char of slack so ANSI-heavy rows
// (swatches) never exactly hit the measured inner boundary.
func SectionBoxInnerWidth(boxWidth int) int {
	if boxWidth < 10 {
		boxWidth = 10
	}
	return max(0, boxWidth-1)
}

// SwatchGridCols picks a column count for swatch grids so rows stay within the
// inner box width on typical terminal sizes (avoids viewport clipping the right border).
func SwatchGridCols(innerW int) int {
	if innerW < 28 {
		return 1
	}
	c := innerW / 32
	if c < 2 {
		return 2
	}
	if c > 4 {
		return 4
	}
	return c
}

// truncateLinesToDisplayWidth trims each line to maxW terminal cells (ANSI-aware)
// so lipgloss and the viewport never clip the right border on overflow lines.
func truncateLinesToDisplayWidth(s string, maxW int) string {
	if maxW <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i := range lines {
		if ansi.StringWidth(lines[i]) > maxW {
			lines[i] = ansi.Cut(lines[i], 0, maxW)
		}
	}
	return strings.Join(lines, "\n")
}

// SectionBox wraps content in a titled bordered lipgloss box.
// width is the inner content width (lipgloss Width() semantics); the rendered
// outer box will be width+4 chars wide (1 border + 1 padding on each side).
// Callers should pass vpWidth-4 so the outer box exactly fills the viewport.
func SectionBox(title, content string, width int) string {
	if width < 10 {
		width = 10
	}
	innerW := SectionBoxInnerWidth(width)
	title = truncateLinesToDisplayWidth(title, innerW)
	content = truncateLinesToDisplayWidth(content, innerW)
	inner := title + "\n" + content
	return sectionBoxStyle.Width(width).Render(inner) + "\n"
}

// hexRe matches bare #rrggbb hex color strings (case-insensitive, word boundary).
var hexRe = regexp.MustCompile(`(?i)#[0-9a-fA-F]{6}`)

// ColorizeHex replaces #rrggbb patterns in text with a colored swatch + the hex string.
func ColorizeHex(text string) string {
	return hexRe.ReplaceAllStringFunc(text, func(h string) string {
		c := hexToCentroid(h)
		return Swatch(c) + h
	})
}

// hexToCentroid parses a #rrggbb string into a Centroid for swatch rendering.
func hexToCentroid(h string) color.Centroid {
	s := strings.TrimPrefix(h, "#")
	if len(s) != 6 {
		return color.Centroid{}
	}
	r, err1 := strconv.ParseUint(s[0:2], 16, 8)
	g, err2 := strconv.ParseUint(s[2:4], 16, 8)
	b, err3 := strconv.ParseUint(s[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return color.Centroid{}
	}
	return color.Centroid{R: uint8(r), G: uint8(g), B: uint8(b)}
}
