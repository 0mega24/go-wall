package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/0mega24/gowall/internal/pipeline"
)

func (m tuiModel) viewPreviewTab() string {
	if m.result == nil {
		return "  No palette loaded yet.\n"
	}
	if m.contentBuilding[tabPreview] && m.renderedContent[tabPreview] == "" {
		return "  building…\n"
	}
	return viewportPlainView(m.previewViewport, m.renderedContent[tabPreview]) + "\n"
}

type previewSnapshot struct {
	result  *pipeline.Result
	vpWidth int
}

func newPreviewSnapshot(m tuiModel) tabBuilder {
	return previewSnapshot{result: m.result, vpWidth: m.previewViewport.Width}
}

func (s previewSnapshot) build() string {
	if s.result == nil {
		return "No palette loaded yet."
	}
	// vpWidth is the viewport outer width; subtract 4 (border+padding on each side)
	// so the SectionBox outer (width+4) exactly fills the viewport.
	boxW := s.vpWidth - 4
	if boxW < 10 {
		boxW = 10
	}
	inner := SectionBoxInnerWidth(boxW)
	cols := SwatchGridCols(inner)
	var b strings.Builder

	var filtBuf strings.Builder
	filtBuf.WriteString(GradientRow(s.result.Filtered, inner))
	filtBuf.WriteString("\n")
	filtBuf.WriteString(RowOfSwatchHex(s.result.Filtered, cols, "  "))
	b.WriteString(SectionBox("Filtered Palette", filtBuf.String(), boxW))

	var ansiBuf strings.Builder
	for i, c := range s.result.ANSI {
		if i > 0 && i%cols == 0 {
			ansiBuf.WriteString("\n")
		} else if i > 0 {
			ansiBuf.WriteString("  ")
		}
		fg := contrastFG(c.R, c.G, c.B)
		style := lipgloss.NewStyle().
			Background(lipgloss.Color(c.Hex())).
			Foreground(lipgloss.Color(fg))
		// Fixed-width cell: index + hex so columns line up for 0–9 vs 10–15.
		label := fmt.Sprintf(" %2d %-7s ", i, c.Hex())
		ansiBuf.WriteString(style.Render(label))
	}
	b.WriteString(SectionBox("ANSI 16 Slots", ansiBuf.String(), boxW))

	var tonesBuf strings.Builder
	tonesBuf.WriteString(GradientRow(s.result.Tones, inner))
	tonesBuf.WriteString("\n")
	tonesBuf.WriteString(RowOfSwatchHex(s.result.Tones, cols, "  "))
	b.WriteString(SectionBox("Tone Ramp (bg → fg)", tonesBuf.String(), boxW))

	if len(s.result.TonesFromANSI) > 0 {
		var tfaBuf strings.Builder
		tfaBuf.WriteString(GradientRow(s.result.TonesFromANSI, inner))
		tfaBuf.WriteString("\n")
		tfaBuf.WriteString(RowOfSwatchHex(s.result.TonesFromANSI, cols, "  "))
		b.WriteString(SectionBox("Tones from ANSI", tfaBuf.String(), boxW))
	}

	if len(s.result.Tones) >= 2 {
		bg := s.result.Tones[0]
		fg := s.result.Tones[len(s.result.Tones)-1]
		sample := lipgloss.NewStyle().
			Background(lipgloss.Color(bg.Hex())).
			Foreground(lipgloss.Color(fg.Hex())).
			Padding(0, 1)
		termBuf := "bg " + SwatchHex(bg) + "   fg " + SwatchHex(fg) + "\n" +
			sample.Render("The quick brown fox jumps over the lazy dog")
		b.WriteString(SectionBox("Terminal Sample", termBuf, boxW))
	}

	var refBuf strings.Builder
	fmt.Fprintf(&refBuf, "Background = #%s   {{ .Background }}\n", s.result.Theme.Background)
	fmt.Fprintf(&refBuf, "Foreground = #%s   {{ .Foreground }}\n", s.result.Theme.Foreground)
	refBuf.WriteString("\n")
	for i, c := range s.result.ANSI {
		fmt.Fprintf(&refBuf, "%s  %s  {{ index .Ansi %2d }}\n", Swatch(c), c.Hex(), i)
	}
	b.WriteString(SectionBox("Full Reference", refBuf.String(), boxW))

	return b.String()
}

// contrastFG returns "#000000" or "#ffffff" based on which provides better contrast.
func contrastFG(r, g, b uint8) string {
	lum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255
	if lum > 0.5 {
		return "#000000"
	}
	return "#ffffff"
}
