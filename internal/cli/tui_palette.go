package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/pipeline"
)

func (m tuiModel) viewPaletteTab() string {
	if m.result == nil {
		return "  No palette loaded yet.\n"
	}
	if m.contentBuilding[tabPalette] && m.renderedContent[tabPalette] == "" {
		var b strings.Builder
		b.WriteString("  building…")
		if m.configMsg != "" {
			b.WriteString("\n  " + m.configMsg)
		}
		return b.String()
	}
	var b strings.Builder
	b.WriteString(viewportPlainView(m.paletteViewport, m.renderedContent[tabPalette]))
	if m.configMsg != "" {
		b.WriteString("\n  " + m.configMsg)
	}
	return b.String()
}

type paletteSnapshot struct {
	result         *pipeline.Result
	swatchMode     int
	vpWidth        int
	manualEditOpen bool
	manualSlot     int
	manualHexBuf   string
	hasManualANSI  bool
}

func newPaletteSnapshot(m tuiModel) tabBuilder {
	w := m.paletteViewport.Width
	if w < 10 {
		w = m.innerW() - 6
		if w < 10 {
			w = 10
		}
	}
	hasManual := false
	for _, p := range m.paletteManualANSI {
		if p != nil {
			hasManual = true
			break
		}
	}
	return paletteSnapshot{
		result:         m.result,
		swatchMode:     m.swatchMode,
		vpWidth:        w,
		manualEditOpen: m.paletteManualEditOpen,
		manualSlot:     m.paletteManualSlot,
		manualHexBuf:   m.paletteManualHexBuf,
		hasManualANSI:  hasManual,
	}
}

func (s paletteSnapshot) build() string {
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

	if s.manualEditOpen {
		b.WriteString(SectionBox("Manual ANSI (unverified)",
			fmt.Sprintf(
				"Not checked for contrast, distinction, or readability.\n"+
					"Slot %d  ([ / ] change)  Hex: %s\n"+
					"Enter apply · esc cancel",
				s.manualSlot, s.manualHexBuf,
			), boxW))
	}

	var filtContent strings.Builder
	if s.swatchMode == 1 {
		filtContent.WriteString(GradientRow(s.result.Filtered, inner))
	} else {
		filtContent.WriteString(rowWithIndent(RowOfSwatchHex(s.result.Filtered, cols, "  ")))
	}
	b.WriteString(SectionBox("Filtered Palette", filtContent.String(), boxW))

	var ansiContent strings.Builder
	if s.swatchMode == 1 {
		ansiContent.WriteString(GradientRow(s.result.ANSI, inner))
	} else {
		ansiContent.WriteString(ansiGridWithVars(s.result.ANSI, cols))
	}
	b.WriteString(SectionBox("ANSI Colors  (.Ansi N)", ansiContent.String(), boxW))

	var tonesContent strings.Builder
	if s.swatchMode == 1 {
		tonesContent.WriteString(GradientRow(s.result.Tones, inner))
	} else {
		tonesContent.WriteString(tonesWithVars(s.result.Tones, "Tones", cols))
	}
	b.WriteString(SectionBox("Tones bg→fg  (.Tones N)", tonesContent.String(), boxW))

	if len(s.result.TonesFromANSI) > 0 {
		var tfaContent strings.Builder
		if s.swatchMode == 1 {
			tfaContent.WriteString(GradientRow(s.result.TonesFromANSI, inner))
		} else {
			tfaContent.WriteString(tonesWithVars(s.result.TonesFromANSI, "TonesFromANSI", cols))
		}
		b.WriteString(SectionBox("Tones from ANSI  (.TonesFromANSI N)", tfaContent.String(), boxW))
	}

	if len(s.result.Tones) >= 2 {
		bg := s.result.Tones[0]
		fg := s.result.Tones[len(s.result.Tones)-1]
		bgfgContent := fmt.Sprintf(
			"Background: %s  .Background = #%s\nForeground: %s  .Foreground = #%s",
			SwatchHex(bg), s.result.Theme.Background,
			SwatchHex(fg), s.result.Theme.Foreground,
		)
		b.WriteString(SectionBox("Background / Foreground", bgfgContent, boxW))
	}

	if s.hasManualANSI && !s.manualEditOpen {
		b.WriteString(SectionBox("Notice", "Manual ANSI overrides are unverified. Press u on this tab to clear and regenerate from the image.", boxW))
	}

	return b.String()
}

func (m tuiModel) handlePaletteKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	if m.paletteManualEditOpen {
		return m.handlePaletteManualKey(msg)
	}
	switch msg.String() {
	case "m":
		m.paletteManualEditOpen = true
		m.paletteManualSlot = 0
		m.paletteManualHexBuf = ""
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		return m, mdCmd
	case "u":
		if !m.hasManualANSIOverrides() {
			return m, nil
		}
		for i := range m.paletteManualANSI {
			m.paletteManualANSI[i] = nil
		}
		m.paletteManualEditOpen = false
		m = m.rememberTabForPipelineRun()
		m.state = "loading"
		return m, m.pipelineCmd()
	case "s":
		m.swatchMode = (m.swatchMode + 1) % 2
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		var cmd tea.Cmd
		m.paletteViewport, cmd = m.paletteViewport.Update(msg)
		return m, tea.Batch(mdCmd, cmd)
	case "e":
		if m.result != nil {
			path := "./gowall-colors.txt"
			if err := exportColorReference(m.result, path); err != nil {
				m.configMsg = "Export failed: " + err.Error()
			} else {
				m.configMsg = "Exported to " + path
			}
		}
	}
	var cmd tea.Cmd
	m.paletteViewport, cmd = m.paletteViewport.Update(msg)
	return m, cmd
}

func (m tuiModel) handlePaletteManualKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.paletteManualEditOpen = false
		m.paletteManualHexBuf = ""
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		return m, mdCmd
	case "enter":
		return m.tryApplyPaletteManualHex()
	case "backspace":
		if len(m.paletteManualHexBuf) > 0 {
			m.paletteManualHexBuf = m.paletteManualHexBuf[:len(m.paletteManualHexBuf)-1]
		}
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		return m, mdCmd
	case "[":
		m.paletteManualSlot = (m.paletteManualSlot + 15) % 16
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		return m, mdCmd
	case "]":
		m.paletteManualSlot = (m.paletteManualSlot + 1) % 16
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		return m, mdCmd
	default:
		if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				if r == '#' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
					if len(m.paletteManualHexBuf) < 7 {
						m.paletteManualHexBuf += string(r)
					}
				}
			}
			var mdCmd tea.Cmd
			m, mdCmd = m.markDirty(tabPalette)
			return m, mdCmd
		}
	}
	var cmd tea.Cmd
	m.paletteViewport, cmd = m.paletteViewport.Update(msg)
	return m, cmd
}

func (m tuiModel) tryApplyPaletteManualHex() (tuiModel, tea.Cmd) {
	h := strings.TrimSpace(strings.TrimPrefix(m.paletteManualHexBuf, "#"))
	if len(h) != 6 {
		m.configMsg = "Manual hex: need 6 hex digits (e.g. #aabbcc)"
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		return m, mdCmd
	}
	c, err := pipeline.ParseHex("#" + h)
	if err != nil {
		m.configMsg = "Invalid hex: " + err.Error()
		var mdCmd tea.Cmd
		m, mdCmd = m.markDirty(tabPalette)
		return m, mdCmd
	}
	slot := m.paletteManualSlot
	if slot < 0 || slot > 15 {
		slot = 0
	}
	c2 := c
	m.paletteManualANSI[slot] = &c2
	m = m.applyManualANSIOverridesToResult()
	m.paletteManualEditOpen = false
	m.paletteManualHexBuf = ""
	m.configMsg = fmt.Sprintf("ANSI %d set (unverified)", slot)
	var cmd tea.Cmd
	m, cmd = m.markAllTabsDirty()
	return m, cmd
}

// rowWithIndent prepends two spaces to a content string for consistent box indentation.
func rowWithIndent(s string) string {
	return "  " + strings.ReplaceAll(s, "\n", "\n  ")
}

// ansiGridWithVars renders ANSI colors in a grid with fixed-width columns so
// single- vs double-digit indices (.Ansi 9 vs .Ansi 10) stay aligned.
func ansiGridWithVars(colors []color.Centroid, perRow int) string {
	var b strings.Builder
	for i, c := range colors {
		if perRow > 0 && i > 0 && i%perRow == 0 {
			b.WriteString("\n")
		} else if i > 0 {
			b.WriteString("  ")
		}
		fmt.Fprintf(&b, "%s  %7s  .Ansi %2d", Swatch(c), c.Hex(), i)
	}
	return b.String()
}

// tonesWithVars renders tone colors with simplified template variable syntax.
func tonesWithVars(colors []color.Centroid, varName string, perRow int) string {
	var b strings.Builder
	for i, c := range colors {
		if perRow > 0 && i > 0 && i%perRow == 0 {
			b.WriteString("\n")
		} else if i > 0 {
			b.WriteString("  ")
		}
		fmt.Fprintf(&b, "%s  %7s  .%s %2d", Swatch(c), c.Hex(), varName, i)
	}
	return b.String()
}

// exportColorReference writes a text file listing all palette colors with template variable syntax.
func exportColorReference(result *pipeline.Result, path string) error {
	var b strings.Builder
	b.WriteString("# Gowall Color Reference\n")
	b.WriteString("# Use in templates as: {{ .Background }}, {{ index .Ansi 3 }}, etc.\n\n")

	fmt.Fprintf(&b, ".Background          = #%s\n", result.Theme.Background)
	fmt.Fprintf(&b, ".Foreground          = #%s\n\n", result.Theme.Foreground)

	b.WriteString("# ANSI slots  ({{ index .Ansi N }})\n")
	for i, c := range result.ANSI {
		fmt.Fprintf(&b, ".Ansi %-2d              = %s\n", i, c.Hex())
	}
	b.WriteString("\n")

	b.WriteString("# Tone ramp bg→fg  ({{ index .Tones N }})\n")
	for i, c := range result.Tones {
		fmt.Fprintf(&b, ".Tones %-2d             = %s\n", i, c.Hex())
	}
	b.WriteString("\n")

	if len(result.TonesFromANSI) > 0 {
		b.WriteString("# Tones from ANSI  ({{ index .TonesFromANSI N }})\n")
		for i, c := range result.TonesFromANSI {
			fmt.Fprintf(&b, ".TonesFromANSI %-2d     = %s\n", i, c.Hex())
		}
		b.WriteString("\n")
	}

	b.WriteString("# Raw filtered palette (no template variable — reference only)\n")
	for i, c := range result.Filtered {
		fmt.Fprintf(&b, "# Filtered[%-2d]         = %s\n", i, c.Hex())
	}

	return os.WriteFile(filepath.Clean(path), []byte(b.String()), 0o600)
}
