package cli

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// contentXOffset is the horizontal offset of the bordered pane when the render
// width is narrower than the terminal (centered with PlaceHorizontal).
func (m tuiModel) contentXOffset() int {
	if m.termW <= m.width {
		return 0
	}
	return (m.termW - m.innerW()) / 2
}

// contentXFromMouse maps a terminal mouse X to 0-based column inside the inner
// pane (first character of base content after the left │).
func (m tuiModel) contentXFromMouse(msg tea.MouseMsg) int {
	return msg.X - m.contentXOffset() - 1
}

func (m tuiModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.state == "picking" {
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)
		return m, cmd
	}
	if m.showHelp {
		return m, m.helpOverlay.Update(msg)
	}
	if !m.showingPreviewUI() {
		return m, nil
	}

	if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
		switch m.tab {
		case tabTemplates:
			var cmd tea.Cmd
			m.tmplViewport, cmd = m.tmplViewport.Update(msg)
			return m, cmd
		case tabPalette:
			var cmd tea.Cmd
			m.paletteViewport, cmd = m.paletteViewport.Update(msg)
			return m, cmd
		case tabPreview:
			var cmd tea.Cmd
			m.previewViewport, cmd = m.previewViewport.Update(msg)
			return m, cmd
		default:
			return m, nil
		}
	}

	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	if nm, cmd, ok := m.handleMouseUILeft(msg); ok {
		return nm, cmd
	}
	return m, nil
}

func (m tuiModel) handleMouseUILeft(msg tea.MouseMsg) (tuiModel, tea.Cmd, bool) {
	lx := m.contentXFromMouse(msg)
	innerContentW := m.innerW() - 2
	if innerContentW < 1 {
		return m, nil, false
	}
	if lx < 0 || lx >= innerContentW {
		return m, nil, false
	}

	// Tab bar middle row (tab labels): Y = 2 with outer border at Y = 0.
	if msg.Y == 2 {
		if tabIdx, ok := m.tabBarHitTabIndex(lx); ok {
			if tabIdx != m.tab {
				m.tab = tabIdx
				m = syncHelpOverlayForPane(m)
			}
			return m, nil, true
		}
		if act, ok := m.tabBarHitHint(lx); ok {
			return m.mouseDispatchTopHint(act)
		}
	}

	base := m.renderBase()
	statusY := lipgloss.Height(base)
	if msg.Y == statusY {
		if act, ok := m.statusBarHit(lx); ok {
			return m.mouseDispatchStatus(act)
		}
	}

	return m, nil, false
}

func (m tuiModel) tabBarHitTabIndex(lx int) (int, bool) {
	names := []string{"Config", "Adjust", "Templates", "Palette", "Preview"}
	x := 0
	for i, name := range names {
		var rendered string
		tabActive := i == m.tab && (m.state != "loading" || m.result != nil)
		if tabActive {
			rendered = activeTabStyle.Render(name)
		} else {
			rendered = inactiveTabStyle.Render(name)
		}
		lines := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")
		for len(lines) < 3 {
			lines = append(lines, "")
		}
		top := lines[0]
		w := lipgloss.Width(top)
		if i > 0 {
			x++
		}
		if lx >= x && lx < x+w {
			return i, true
		}
		x += w
	}
	return 0, false
}

func (m tuiModel) tabBarHitHint(lx int) (string, bool) {
	hintText := "? help  o open  I import  q quit  "
	hintStyled := tabHintFaintStyle.Render("  " + hintText)
	hintW := lipgloss.Width(hintStyled)

	names := []string{"Config", "Adjust", "Templates", "Palette", "Preview"}
	var midB strings.Builder
	totalTabsWidth := 0
	for i, name := range names {
		var rendered string
		tabActive := i == m.tab && (m.state != "loading" || m.result != nil)
		if tabActive {
			rendered = activeTabStyle.Render(name)
		} else {
			rendered = inactiveTabStyle.Render(name)
		}
		lines := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")
		for len(lines) < 3 {
			lines = append(lines, "")
		}
		top, mid, _ := lines[0], lines[1], lines[2]
		if i > 0 {
			midB.WriteByte(' ')
			totalTabsWidth++
		}
		midB.WriteString(mid)
		totalTabsWidth += lipgloss.Width(top)
	}

	fillW := m.innerW() - totalTabsWidth - hintW
	if fillW < 0 {
		fillW = 0
	}
	hintStart := lipgloss.Width(midB.String()) + fillW
	if lx < hintStart || lx >= hintStart+hintW {
		return "", false
	}
	rel := lx - hintStart

	plain := "  " + hintText
	for i := 1; i <= len(plain); i++ {
		prefix := plain[:i]
		w := lipgloss.Width(tabHintFaintStyle.Render(prefix))
		if w > rel {
			off := i - 1
			if off < 2 {
				return "", false
			}
			return hintActionForPlainOffset(hintText, off-2)
		}
	}
	return "", false
}

func hintActionForPlainOffset(hintText string, off int) (string, bool) {
	if off < 0 || off >= len(hintText) {
		return "", false
	}
	if i := strings.Index(hintText, "? help"); i >= 0 && off >= i && off < i+len("? help") {
		return "help", true
	}
	if i := strings.Index(hintText, "o open"); i >= 0 && off >= i && off < i+len("o open") {
		return "open", true
	}
	if i := strings.Index(hintText, "I import"); i >= 0 && off >= i && off < i+len("I import") {
		return "import", true
	}
	if i := strings.Index(hintText, "q quit"); i >= 0 && off >= i && off < i+len("q quit") {
		return "quit", true
	}
	return "", false
}

func (m tuiModel) statusBarHit(lx int) (string, bool) {
	imgName := "(no image)"
	if m.imagePath != "" {
		parts := strings.Split(m.imagePath, "/")
		imgName = parts[len(parts)-1]
	}
	algo := orderedAlgorithms[m.algorithmIdx]
	left := fmt.Sprintf("  %s · %s · k=%d", imgName, algo, m.kVal)
	if m.hasManualANSIOverrides() {
		left += " · unverified ANSI"
	}
	tabNames := []string{"Config", "Adjust", "Templates", "Palette", "Preview"}
	tabLabel := ""
	if m.tab >= 0 && m.tab < len(tabNames) {
		tabLabel = "[" + tabNames[m.tab] + "]"
	}
	right := tabLabel + "  "
	if m.state == "loading" && m.result != nil {
		right += m.spinner.View() + " running…  "
	} else {
		switch m.tab {
		case tabPalette:
			right += "↑↓ scroll · m manual · u clear · Tab · q  "
		case tabPreview:
			right += "↑↓ PgUp/Dn · Tab · q  "
		}
	}

	leftStr := statusBarBaseStyle.Render(left)
	rightStr := statusBarBaseStyle.Render(right)
	fillW := m.innerW() - lipgloss.Width(leftStr) - lipgloss.Width(rightStr)
	if fillW < 0 {
		fillW = 0
	}
	rightStart := lipgloss.Width(leftStr) + fillW
	if lx < rightStart {
		return "", false
	}
	rel := lx - rightStart
	if rel < 0 || rel >= lipgloss.Width(rightStr) {
		return "", false
	}

	// Map column in styled right bar to byte index in plain `right`
	clickIdx := -1
	for i := 1; i <= len(right); i++ {
		prefix := statusBarBaseStyle.Render(right[:i])
		if lipgloss.Width(prefix) > rel {
			clickIdx = i - 1
			break
		}
	}
	if clickIdx < 0 {
		return "", false
	}

	// Current tab label in brackets — switch tab by clicking the name
	if len(tabLabel) > 0 && clickIdx < len(tabLabel) {
		for ti, tn := range tabNames {
			label := "[" + tn + "]"
			if strings.HasPrefix(right, label) && clickIdx >= 0 && clickIdx < len(label) {
				return "status_tab_" + strconv.Itoa(ti), true
			}
		}
	}

	if m.state == "loading" && m.result != nil {
		return "", false
	}

	prefix := tabLabel + "  "
	if clickIdx < len(prefix) {
		return "", false
	}
	plainRest := right[len(prefix):]
	rel2 := clickIdx - len(prefix)
	if rel2 < 0 || rel2 >= len(plainRest) {
		return "", false
	}

	switch m.tab {
	case tabPalette:
		if i := strings.Index(plainRest, "m manual"); i >= 0 && rel2 >= i && rel2 < i+len("m manual") {
			return "palette_m", true
		}
		if i := strings.Index(plainRest, "u clear"); i >= 0 && rel2 >= i && rel2 < i+len("u clear") {
			return "palette_u", true
		}
		if i := strings.Index(plainRest, "Tab"); i >= 0 && rel2 >= i && rel2 < i+len("Tab") {
			return "tab_next", true
		}
	case tabPreview:
		if i := strings.Index(plainRest, "Tab"); i >= 0 && rel2 >= i && rel2 < i+len("Tab") {
			return "tab_next", true
		}
	}
	if j := strings.LastIndex(strings.TrimRight(plainRest, " "), "q"); j >= 0 && rel2 == j {
		return "quit", true
	}
	return "", false
}

func (m tuiModel) mouseDispatchTopHint(act string) (tuiModel, tea.Cmd, bool) {
	switch act {
	case "help":
		if m.showingPreviewUI() {
			m.showHelp = true
			m = syncHelpOverlayForPane(m)
		}
		return m, nil, true
	case "open":
		mod, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
		return mod.(tuiModel), cmd, true
	case "import":
		mod, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
		return mod.(tuiModel), cmd, true
	case "quit":
		mod, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		return mod.(tuiModel), cmd, true
	default:
		return m, nil, false
	}
}

func (m tuiModel) mouseDispatchStatus(act string) (tuiModel, tea.Cmd, bool) {
	switch act {
	case "tab_next":
		mod, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
		return mod.(tuiModel), cmd, true
	case "palette_m":
		nm, cmd := m.handlePaletteKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
		return nm, cmd, true
	case "palette_u":
		nm, cmd := m.handlePaletteKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
		return nm, cmd, true
	case "quit":
		mod, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		return mod.(tuiModel), cmd, true
	}
	if strings.HasPrefix(act, "status_tab_") {
		ti, err := strconv.Atoi(act[len("status_tab_"):])
		if err != nil || ti < 0 || ti >= numTabs {
			return m, nil, false
		}
		if ti != m.tab {
			m.tab = ti
			m = syncHelpOverlayForPane(m)
		}
		return m, nil, true
	}
	return m, nil, false
}
