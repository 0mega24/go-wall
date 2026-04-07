package cli

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/pipeline"
)

var (
	adjustSectionActiveStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	adjustSectionDimStyle    = lipgloss.NewStyle().Faint(true)
	adjustPendingStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
)

func adjustNeedsLiveView(m tuiModel) bool {
	if m.adjustSection == 2 {
		return true
	}
	if m.adjustSection == 3 {
		return true
	}
	if m.adjustSection == 0 && m.slotMode[m.activeSlot] == "pin" {
		return true
	}
	return false
}

func (m tuiModel) viewAdjustTab() string {
	if m.result == nil {
		return "  No palette loaded yet.\n"
	}
	if adjustNeedsLiveView(m) {
		return m.viewAdjustTabLive()
	}
	if m.contentBuilding[tabAdjust] && m.renderedContent[tabAdjust] == "" {
		return "  building…\n"
	}
	if m.renderedContent[tabAdjust] != "" {
		return m.renderedContent[tabAdjust]
	}
	return m.viewAdjustTabLive()
}

func (m tuiModel) viewAdjustTabLive() string {
	if m.result == nil {
		return "  No palette loaded yet.\n"
	}

	w := m.innerW() - 4
	if w < 40 {
		w = 40
	}
	innerW := w - 4

	sectionNames := []string{"ANSI Slots", "Tones", "BG / FG", "Global HSV"}
	var sectionNav strings.Builder
	for i, n := range sectionNames {
		if i > 0 {
			sectionNav.WriteString("  ")
		}
		if i == m.adjustSection {
			sectionNav.WriteString(adjustSectionActiveStyle.Render("[" + n + "]"))
		} else {
			sectionNav.WriteString(adjustSectionDimStyle.Render(" " + n + " "))
		}
	}

	switch m.adjustSection {
	case 0:
		return m.viewAdjustANSI(innerW, sectionNav.String())
	case 1:
		return m.viewAdjustTones(innerW, sectionNav.String())
	case 2:
		return m.viewAdjustBGFG(innerW, sectionNav.String())
	case 3:
		return m.viewAdjustGlobal(innerW, sectionNav.String())
	}
	return "  No palette loaded yet.\n"
}

type adjustSnapshot struct {
	result              *pipeline.Result
	slotMode            [16]string
	lockH               [16]float64
	lockS               [16]float64
	lockV               [16]float64
	hexInputValues      [16]string
	bgInputVal          string
	activeSlot          int
	adjustSection       int
	adjustPending       bool
	globalAdjust        pipeline.GlobalAdjust
	globalAdjustPending bool
	width               int
}

func newAdjustSnapshot(m tuiModel) tabBuilder {
	return adjustSnapshot{
		result:              m.result,
		slotMode:            m.slotMode,
		lockH:               m.lockH,
		lockS:               m.lockS,
		lockV:               m.lockV,
		hexInputValues:      m.hexInputValues,
		bgInputVal:          m.bgInputVal,
		activeSlot:          m.activeSlot,
		adjustSection:       m.adjustSection,
		adjustPending:       m.adjustPending,
		globalAdjust:        m.globalAdjust,
		globalAdjustPending: m.globalAdjustPending,
		width:               m.innerW(),
	}
}

func (s adjustSnapshot) build() string {
	if s.result == nil {
		return "  No palette loaded yet.\n"
	}
	w := s.width - 4
	if w < 40 {
		w = 40
	}
	innerW := w - 4

	sectionNames := []string{"ANSI Slots", "Tones", "BG / FG", "Global HSV"}
	var sectionNav strings.Builder
	for i, n := range sectionNames {
		if i > 0 {
			sectionNav.WriteString("  ")
		}
		if i == s.adjustSection {
			sectionNav.WriteString(adjustSectionActiveStyle.Render("[" + n + "]"))
		} else {
			sectionNav.WriteString(adjustSectionDimStyle.Render(" " + n + " "))
		}
	}

	switch s.adjustSection {
	case 0:
		return s.viewAdjustANSI(innerW, sectionNav.String())
	case 1:
		return s.viewAdjustTones(innerW, sectionNav.String())
	case 2:
		return s.viewAdjustBGFG(innerW, sectionNav.String())
	case 3:
		return s.viewAdjustGlobal(innerW, sectionNav.String())
	}
	return "  No palette loaded yet.\n"
}

func (s adjustSnapshot) viewAdjustANSI(innerW int, sectionNav string) string {
	var gridBuf strings.Builder
	for i, c := range s.result.ANSI {
		if i > 0 && i%4 == 0 {
			gridBuf.WriteString("\n")
		} else if i > 0 {
			gridBuf.WriteString(" ")
		}
		fg := contrastFG(c.R, c.G, c.B)
		style := lipgloss.NewStyle().
			Background(lipgloss.Color(c.Hex())).
			Foreground(lipgloss.Color(fg))
		if i == s.activeSlot {
			style = style.Bold(true).Underline(true)
		}
		gridBuf.WriteString(style.Render(fmt.Sprintf("[%2d]", i)))
	}
	gridTitle := fmt.Sprintf("%s  ·  ↑↓←→ navigate  1/2/3/4 section", sectionNav)
	gridBox := SectionBox(gridTitle, gridBuf.String(), innerW)

	slot := s.activeSlot
	var detailBuf strings.Builder

	if slot < len(s.result.ANSI) {
		c := s.result.ANSI[slot]
		fmt.Fprintf(&detailBuf, "Color: %s  %s\n", Swatch(c), c.Hex())
	}

	mode := s.slotMode[slot]
	pendingTag := ""
	if s.adjustPending && mode == "lock" {
		pendingTag = "  " + adjustPendingStyle.Render("* unsaved — Enter/s to apply")
	}
	fmt.Fprintf(&detailBuf, "Mode:  [%s]  (m to cycle: auto→lock→pin→auto  esc/r reset)%s\n", mode, pendingTag)

	switch mode {
	case "lock":
		fmt.Fprintf(&detailBuf, "H: %5.1f°  (← / → adjust)\n", s.lockH[slot])
		fmt.Fprintf(&detailBuf, "S: %5.3f   ([ / ] adjust)\n", s.lockS[slot])
		fmt.Fprintf(&detailBuf, "V: %5.3f   (- / + adjust)\n", s.lockV[slot])
	case "pin":
		fmt.Fprintf(&detailBuf, "Hex: %s\n  Enter to apply", s.hexInputValues[slot])
	}

	detailBox := SectionBox(fmt.Sprintf("Slot %d", slot), detailBuf.String(), innerW)

	return gridBox + "\n" + detailBox + "\n"
}

func (s adjustSnapshot) viewAdjustTones(innerW int, sectionNav string) string {
	var buf strings.Builder
	for i, c := range s.result.Tones {
		cursor := "  "
		if i == s.activeSlot {
			cursor = "▸ "
		}
		fmt.Fprintf(&buf, "%s%2d  %s  %s\n", cursor, i, Swatch(c), c.Hex())
	}
	title := fmt.Sprintf("%s  ·  ↑↓ navigate  p pin hex  esc reset  1/2/3/4 section", sectionNav)
	box := SectionBox(title, strings.TrimRight(buf.String(), "\n"), innerW)
	return box + "\n"
}

func (s adjustSnapshot) viewAdjustBGFG(innerW int, sectionNav string) string {
	var buf strings.Builder
	if len(s.result.Tones) >= 2 {
		bg := s.result.Tones[0]
		fg := s.result.Tones[len(s.result.Tones)-1]
		fmt.Fprintf(&buf, "Background:  %s  #%s\n", Swatch(bg), s.result.Theme.Background)
		fmt.Fprintf(&buf, "Foreground:  %s  #%s\n", Swatch(fg), s.result.Theme.Foreground)
		buf.WriteString("\nOverride BG hex (empty = auto):\n  ")
		if s.bgInputVal == "" {
			buf.WriteString("(empty)")
		} else {
			buf.WriteString(s.bgInputVal)
		}
		buf.WriteString("\n\n  Enter to apply")
	}
	title := fmt.Sprintf("%s  ·  Enter apply  esc cancel  1/2/3/4 section", sectionNav)
	box := SectionBox(title, buf.String(), innerW)
	return box + "\n"
}

func (m tuiModel) viewAdjustANSI(innerW int, sectionNav string) string {
	var gridBuf strings.Builder
	for i, c := range m.result.ANSI {
		if i > 0 && i%4 == 0 {
			gridBuf.WriteString("\n")
		} else if i > 0 {
			gridBuf.WriteString(" ")
		}
		fg := contrastFG(c.R, c.G, c.B)
		style := lipgloss.NewStyle().
			Background(lipgloss.Color(c.Hex())).
			Foreground(lipgloss.Color(fg))
		if i == m.activeSlot {
			style = style.Bold(true).Underline(true)
		}
		gridBuf.WriteString(style.Render(fmt.Sprintf("[%2d]", i)))
	}
	gridTitle := fmt.Sprintf("%s  ·  ↑↓←→ navigate  1/2/3/4 section", sectionNav)
	gridBox := SectionBox(gridTitle, gridBuf.String(), innerW)

	slot := m.activeSlot
	var detailBuf strings.Builder

	if slot < len(m.result.ANSI) {
		c := m.result.ANSI[slot]
		fmt.Fprintf(&detailBuf, "Color: %s  %s\n", Swatch(c), c.Hex())
	}

	mode := m.slotMode[slot]
	pendingTag := ""
	if m.adjustPending && mode == "lock" {
		pendingTag = "  " + adjustPendingStyle.Render("* unsaved — Enter/s to apply")
	}
	fmt.Fprintf(&detailBuf, "Mode:  [%s]  (m to cycle: auto→lock→pin→auto  esc/r reset)%s\n", mode, pendingTag)

	switch mode {
	case "lock":
		fmt.Fprintf(&detailBuf, "H: %5.1f°  (← / → adjust)\n", m.lockH[slot])
		fmt.Fprintf(&detailBuf, "S: %5.3f   ([ / ] adjust)\n", m.lockS[slot])
		fmt.Fprintf(&detailBuf, "V: %5.3f   (- / + adjust)\n", m.lockV[slot])
	case "pin":
		detailBuf.WriteString("Hex: ")
		detailBuf.WriteString(m.hexInputs[slot].View())
		detailBuf.WriteString("\n  Enter to apply")
	}

	detailBox := SectionBox(fmt.Sprintf("Slot %d", slot), detailBuf.String(), innerW)

	return gridBox + "\n" + detailBox + "\n"
}

func (m tuiModel) viewAdjustTones(innerW int, sectionNav string) string {
	var buf strings.Builder
	for i, c := range m.result.Tones {
		cursor := "  "
		if i == m.activeSlot {
			cursor = "▸ "
		}
		fmt.Fprintf(&buf, "%s%2d  %s  %s\n", cursor, i, Swatch(c), c.Hex())
	}
	title := fmt.Sprintf("%s  ·  ↑↓ navigate  p pin hex  esc reset  1/2/3/4 section", sectionNav)
	box := SectionBox(title, strings.TrimRight(buf.String(), "\n"), innerW)
	return box + "\n"
}

func (m tuiModel) viewAdjustBGFG(innerW int, sectionNav string) string {
	var buf strings.Builder
	if len(m.result.Tones) >= 2 {
		bg := m.result.Tones[0]
		fg := m.result.Tones[len(m.result.Tones)-1]
		fmt.Fprintf(&buf, "Background:  %s  #%s\n", Swatch(bg), m.result.Theme.Background)
		fmt.Fprintf(&buf, "Foreground:  %s  #%s\n", Swatch(fg), m.result.Theme.Foreground)
		buf.WriteString("\nOverride BG hex (empty = auto):\n")
		buf.WriteString("  " + m.bgInput.View())
		buf.WriteString("\n\n  Enter to apply")
	}
	title := fmt.Sprintf("%s  ·  Enter apply  esc cancel  1/2/3/4 section", sectionNav)
	box := SectionBox(title, buf.String(), innerW)
	return box + "\n"
}

func renderAdjustGlobalSection(innerW int, sectionNav string, ga pipeline.GlobalAdjust, pending bool) string {
	var buf strings.Builder
	pendingTag := ""
	if pending {
		pendingTag = "  " + adjustPendingStyle.Render("* unsaved — Enter to apply")
	}
	buf.WriteString("Global adjust runs on every ANSI color before per-slot locks: hue in degrees; Sat/Val are %% multipliers (S×(1+Sat%%/100), V×(1+Val%%/100)).")
	buf.WriteString(pendingTag)
	buf.WriteString("\n\n")
	fmt.Fprintf(&buf, "Hue: %7.1f deg   (← / →)\n", ga.HueDeg)
	fmt.Fprintf(&buf, "Sat: %+7.1f %%   ([ / ]  negative desaturates relative to each color)\n", ga.SatPct)
	fmt.Fprintf(&buf, "Val: %+7.1f %%   (- / +  darkens / brightens relative to each color)\n", ga.ValPct)
	buf.WriteString("\nReadability and slot spread run after this step so the theme stays usable (may soften strong tweaks).")
	buf.WriteString("\nEnter or s to regenerate theme  ·  Esc / r reset to 0")
	title := fmt.Sprintf("%s  ·  1/2/3/4 section", sectionNav)
	box := SectionBox(title, strings.TrimRight(buf.String(), "\n"), innerW)
	return box + "\n"
}

func (s adjustSnapshot) viewAdjustGlobal(innerW int, sectionNav string) string {
	return renderAdjustGlobalSection(innerW, sectionNav, s.globalAdjust, s.globalAdjustPending)
}

func (m tuiModel) viewAdjustGlobal(innerW int, sectionNav string) string {
	return renderAdjustGlobalSection(innerW, sectionNav, m.globalAdjust, m.globalAdjustPending)
}

func (m tuiModel) handleAdjustGlobalKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	const pctStep = 2.0
	switch msg.String() {
	case "left", "h":
		m.globalAdjust.HueDeg -= 0.5
		m.globalAdjustPending = true
		return m.debounceAdjustDirty()
	case "right", "l":
		m.globalAdjust.HueDeg += 0.5
		m.globalAdjustPending = true
		return m.debounceAdjustDirty()
	case "[":
		m.globalAdjust.SatPct = clampAdj(m.globalAdjust.SatPct-pctStep, -100, 100)
		m.globalAdjustPending = true
		return m.debounceAdjustDirty()
	case "]":
		m.globalAdjust.SatPct = clampAdj(m.globalAdjust.SatPct+pctStep, -100, 100)
		m.globalAdjustPending = true
		return m.debounceAdjustDirty()
	case "-":
		m.globalAdjust.ValPct = clampAdj(m.globalAdjust.ValPct-pctStep, -100, 100)
		m.globalAdjustPending = true
		return m.debounceAdjustDirty()
	case "=", "+":
		m.globalAdjust.ValPct = clampAdj(m.globalAdjust.ValPct+pctStep, -100, 100)
		m.globalAdjustPending = true
		return m.debounceAdjustDirty()
	case "enter", "s":
		m.globalAdjustPending = false
		m = m.rememberTabForPipelineRun()
		m.state = "loading"
		return m, m.pipelineCmd()
	case "r", "esc":
		m.globalAdjust = pipeline.GlobalAdjust{}
		m.globalAdjustPending = false
		m = m.rememberTabForPipelineRun()
		m.state = "loading"
		return m, m.pipelineCmd()
	}
	return m, nil
}

func (m tuiModel) handleAdjustKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	switch msg.String() {
	case "1":
		if m.adjustSection == 2 {
			m.bgInput.Blur()
		}
		m.adjustSection = 0
		return m.debounceAdjustDirty()
	case "2":
		if m.adjustSection == 2 {
			m.bgInput.Blur()
		}
		m.adjustSection = 1
		m.activeSlot = 0
		return m.debounceAdjustDirty()
	case "3":
		m.adjustSection = 2
		_ = m.bgInput.Focus()
		return m, nil
	case "4":
		if m.adjustSection == 2 {
			m.bgInput.Blur()
		}
		m.adjustSection = 3
		return m.debounceAdjustDirty()
	}

	switch m.adjustSection {
	case 1:
		return m.handleAdjustTonesKey(msg)
	case 2:
		return m.handleAdjustBGFGKey(msg)
	case 3:
		return m.handleAdjustGlobalKey(msg)
	}

	return m.handleAdjustANSIKey(msg)
}

func (m tuiModel) handleAdjustANSIKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	slot := m.activeSlot
	mode := m.slotMode[slot]

	if mode == "pin" {
		switch msg.String() {
		case "up", "down", "tab", "shift+tab", "m", "esc", "r":
		case "enter":
			hexStr := m.hexInputValues[slot]
			if c, err := pipeline.ParseHex(hexStr); err == nil {
				if m.constraints == nil {
					m.constraints = make(map[int]pipeline.SlotConstraint)
				}
				sc := m.constraints[slot]
				sc.Pin = &c
				m.constraints[slot] = sc
				m = m.rememberTabForPipelineRun()
				m.state = "loading"
				return m, m.pipelineCmd()
			}
			return m, nil
		default:
			inp := m.hexInputs[slot]
			newModel, cmd := inp.Update(msg)
			if updated, ok := newModel.(*huh.Input); ok {
				m.hexInputs[slot] = updated
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "up", "k":
		if m.activeSlot >= 4 {
			m.activeSlot -= 4
			return m.debounceAdjustDirty()
		}
	case "down", "j":
		if m.activeSlot < 12 {
			m.activeSlot += 4
			return m.debounceAdjustDirty()
		}
	case "left", "h":
		if mode == "lock" {
			m.lockH[slot] = math.Mod(m.lockH[slot]-0.5+360, 360)
			m.adjustPending = true
			return m.debounceAdjustDirty()
		}
		if m.activeSlot%4 > 0 {
			m.activeSlot--
			return m.debounceAdjustDirty()
		}
	case "right", "l":
		if mode == "lock" {
			m.lockH[slot] = math.Mod(m.lockH[slot]+0.5, 360)
			m.adjustPending = true
			return m.debounceAdjustDirty()
		}
		if m.activeSlot%4 < 3 {
			m.activeSlot++
			return m.debounceAdjustDirty()
		}

	case "[":
		if mode == "lock" {
			m.lockS[slot] = clampAdj(m.lockS[slot]-0.01, 0, 1)
			m.adjustPending = true
			return m.debounceAdjustDirty()
		}
	case "]":
		if mode == "lock" {
			m.lockS[slot] = clampAdj(m.lockS[slot]+0.01, 0, 1)
			m.adjustPending = true
			return m.debounceAdjustDirty()
		}
	case "-":
		if mode == "lock" {
			m.lockV[slot] = clampAdj(m.lockV[slot]-0.01, 0, 1)
			m.adjustPending = true
			return m.debounceAdjustDirty()
		}
	case "=", "+":
		if mode == "lock" {
			m.lockV[slot] = clampAdj(m.lockV[slot]+0.01, 0, 1)
			m.adjustPending = true
			return m.debounceAdjustDirty()
		}

	case "enter", "s":
		if mode == "lock" && m.adjustPending {
			m = m.withLockConstraint(slot)
			m.adjustPending = false
			m = m.rememberTabForPipelineRun()
			m.state = "loading"
			return m, m.pipelineCmd()
		}

	case "m":
		switch mode {
		case "auto":
			m.slotMode[slot] = "lock"
			if m.result != nil && slot < len(m.result.ANSI) {
				h, s, v := rgbToHSV(m.result.ANSI[slot])
				m.lockH[slot] = h
				m.lockS[slot] = s
				m.lockV[slot] = v
			}
			m = m.withLockConstraint(slot)
			m.adjustPending = false
			m = m.rememberTabForPipelineRun()
			m.state = "loading"
			return m, m.pipelineCmd()
		case "lock":
			m.slotMode[slot] = "pin"
			m.adjustPending = false
			if m.result != nil && slot < len(m.result.ANSI) {
				m.hexInputValues[slot] = m.result.ANSI[slot].Hex()
			}
			_ = m.hexInputs[slot].Focus()
		case "pin":
			m.slotMode[slot] = "auto"
			if m.constraints != nil {
				delete(m.constraints, slot)
			}
			m = m.rememberTabForPipelineRun()
			m.state = "loading"
			return m, m.pipelineCmd()
		}

	case "r", "esc":
		m.slotMode[slot] = "auto"
		m.adjustPending = false
		if m.constraints != nil {
			delete(m.constraints, slot)
		}
		if mode != "auto" {
			m = m.rememberTabForPipelineRun()
			m.state = "loading"
			return m, m.pipelineCmd()
		}
	}

	return m, nil
}

func (m tuiModel) handleAdjustTonesKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	maxSlot := len(m.result.Tones) - 1
	if maxSlot < 0 {
		return m, nil
	}
	switch msg.String() {
	case "up", "k":
		if m.activeSlot > 0 {
			m.activeSlot--
			return m.debounceAdjustDirty()
		}
	case "down", "j":
		if m.activeSlot < maxSlot {
			m.activeSlot++
			return m.debounceAdjustDirty()
		}
	case "p":
	case "esc", "r":
	}
	return m, nil
}

func (m tuiModel) handleAdjustBGFGKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	switch msg.String() {
	case "up", "down", "tab", "shift+tab", "enter", "r", "esc":
	default:
		newInp, cmd := m.bgInput.Update(msg)
		if updated, ok := newInp.(*huh.Input); ok {
			m.bgInput = updated
		}
		return m, cmd
	}

	switch msg.String() {
	case "enter":
		m = m.rememberTabForPipelineRun()
		m.state = "loading"
		return m, m.pipelineCmd()
	case "esc", "r":
		m.bgInputVal = ""
	}
	return m, nil
}

func (m tuiModel) withLockConstraint(slot int) tuiModel {
	if m.constraints == nil {
		m.constraints = make(map[int]pipeline.SlotConstraint)
	}
	h := m.lockH[slot]
	s := m.lockS[slot]
	v := m.lockV[slot]
	sc := m.constraints[slot]
	sc.Pin = nil
	sc.LockH = &h
	sc.LockS = &s
	sc.LockV = &v
	m.constraints[slot] = sc
	return m
}

func clampAdj(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func rgbToHSV(c color.Centroid) (h, s, v float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	v = max
	if max == 0 {
		s = 0
	} else {
		s = delta / max
	}

	if delta == 0 {
		h = 0
		return h, s, v
	}
	switch max {
	case r:
		h = 60 * math.Mod((g-b)/delta, 6)
	case g:
		h = 60 * ((b-r)/delta + 2)
	case b:
		h = 60 * ((r-g)/delta + 4)
	}
	if h < 0 {
		h += 360
	}
	return h, s, v
}
