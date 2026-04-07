package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	cfgFieldImage     = 0
	cfgFieldAlgorithm = 1
	cfgFieldSeed      = 2
	cfgFieldK         = 3
	cfgFieldIters     = 4
	cfgFieldRetone    = 5
	cfgFieldBg        = 6
	cfgNumFields      = 7
)

var (
	configBoxBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#555555")).
				Padding(0, 1)

	configHintMutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa"))
)

// focusConfigField focuses the huh.Input for the current configField (if any).
func (m tuiModel) focusConfigField() tuiModel {
	switch m.configField {
	case cfgFieldSeed:
		_ = m.seedInput.Focus()
	case cfgFieldBg:
		_ = m.bgInput.Focus()
	}
	return m
}

// blurConfigField blurs the huh.Input for the current configField (if any).
func (m tuiModel) blurConfigField() tuiModel {
	switch m.configField {
	case cfgFieldSeed:
		m.seedInput.Blur()
	case cfgFieldBg:
		m.bgInput.Blur()
	}
	return m
}

type configSnapshot struct {
	imagePath    string
	algorithmIdx int
	seedInputVal string
	kVal         int
	itersVal     int
	retoneANSI   bool
	bgInputVal   string
	configField  int
	configMsg    string
	width        int
}

func newConfigSnapshot(m tuiModel) tabBuilder {
	return configSnapshot{
		imagePath:    m.imagePath,
		algorithmIdx: m.algorithmIdx,
		seedInputVal: m.seedInputVal,
		kVal:         m.kVal,
		itersVal:     m.itersVal,
		retoneANSI:   m.retoneANSI,
		bgInputVal:   m.bgInputVal,
		configField:  m.configField,
		configMsg:    m.configMsg,
		width:        m.innerW() - 4,
	}
}

func (s configSnapshot) build() string {
	w := s.width
	if w < 40 {
		w = 40
	}

	var rows []string
	labelW := 14

	for i := 0; i < cfgNumFields; i++ {
		active := i == s.configField
		prefix := "  "
		if active {
			prefix = "▸ "
		}
		rowStyle := lipgloss.NewStyle()
		if active {
			rowStyle = rowStyle.Bold(true)
		}

		var line string
		switch i {
		case cfgFieldImage:
			imgPath := s.imagePath
			if imgPath == "" {
				imgPath = "(none)"
			}
			maxLen := w - labelW - 15
			if maxLen > 10 && len(imgPath) > maxLen {
				imgPath = "…" + imgPath[len(imgPath)-maxLen+1:]
			}
			line = fmt.Sprintf("%-*s %s  [o open]", labelW, "Image:", imgPath)

		case cfgFieldAlgorithm:
			algo := orderedAlgorithms[s.algorithmIdx]
			line = fmt.Sprintf("%-*s ◀ %-10s ▶  ←→ to change", labelW, "Algorithm:", algo)

		case cfgFieldSeed:
			line = fmt.Sprintf("%-*s %s", labelW, "Seed:", s.seedInputVal)

		case cfgFieldK:
			line = fmt.Sprintf("%-*s %-6d  -/+ adjust  (range 4–256)", labelW, "Clusters:", s.kVal)

		case cfgFieldIters:
			line = fmt.Sprintf("%-*s %-6d  -/+ adjust  (range 1–100)", labelW, "Iterations:", s.itersVal)

		case cfgFieldRetone:
			check := "[ ]"
			if s.retoneANSI {
				check = "[x]"
			}
			line = fmt.Sprintf("%-*s %s  space to toggle", labelW, "Retone ANSI:", check)

		case cfgFieldBg:
			val := s.bgInputVal
			if val == "" {
				val = "(auto)"
			}
			line = fmt.Sprintf("%-*s %s", labelW, "Background:", val)
		}

		rows = append(rows, rowStyle.Render(prefix+line))
	}

	content := strings.Join(rows, "\n")

	var hint string
	if s.configMsg != "" {
		hint = "\n  " + configHintMutedStyle.Render(s.configMsg)
	}

	box := configBoxBorderStyle.
		Width(w).
		Render("Pipeline Config\n\n" + content + "\n\n  ↑↓ move  Enter re-run  r re-run")

	return box + hint + "\n"
}

func (m tuiModel) viewConfigTab() string {
	if m.configField == cfgFieldSeed || m.configField == cfgFieldBg {
		return m.viewConfigTabLive()
	}
	if m.contentBuilding[tabConfig] && m.renderedContent[tabConfig] == "" {
		return "  building…\n"
	}
	if m.renderedContent[tabConfig] != "" {
		return m.renderedContent[tabConfig]
	}
	return m.viewConfigTabLive()
}

func (m tuiModel) viewConfigTabLive() string {
	w := m.innerW() - 4
	if w < 40 {
		w = 40
	}

	var rows []string
	labelW := 14

	for i := 0; i < cfgNumFields; i++ {
		active := i == m.configField
		prefix := "  "
		if active {
			prefix = "▸ "
		}
		rowStyle := lipgloss.NewStyle()
		if active {
			rowStyle = rowStyle.Bold(true)
		}

		var line string
		switch i {
		case cfgFieldImage:
			imgPath := m.imagePath
			if imgPath == "" {
				imgPath = "(none)"
			}
			maxLen := w - labelW - 15
			if maxLen > 10 && len(imgPath) > maxLen {
				imgPath = "…" + imgPath[len(imgPath)-maxLen+1:]
			}
			line = fmt.Sprintf("%-*s %s  [o open]", labelW, "Image:", imgPath)

		case cfgFieldAlgorithm:
			algo := orderedAlgorithms[m.algorithmIdx]
			line = fmt.Sprintf("%-*s ◀ %-10s ▶  ←→ to change", labelW, "Algorithm:", algo)

		case cfgFieldSeed:
			if active {
				line = fmt.Sprintf("%-*s %s", labelW, "Seed:", m.seedInput.View())
			} else {
				line = fmt.Sprintf("%-*s %s", labelW, "Seed:", m.seedInputVal)
			}

		case cfgFieldK:
			line = fmt.Sprintf("%-*s %-6d  -/+ adjust  (range 4–256)", labelW, "Clusters:", m.kVal)

		case cfgFieldIters:
			line = fmt.Sprintf("%-*s %-6d  -/+ adjust  (range 1–100)", labelW, "Iterations:", m.itersVal)

		case cfgFieldRetone:
			check := "[ ]"
			if m.retoneANSI {
				check = "[x]"
			}
			line = fmt.Sprintf("%-*s %s  space to toggle", labelW, "Retone ANSI:", check)

		case cfgFieldBg:
			if active {
				line = fmt.Sprintf("%-*s %s", labelW, "Background:", m.bgInput.View())
			} else {
				val := m.bgInputVal
				if val == "" {
					val = "(auto)"
				}
				line = fmt.Sprintf("%-*s %s", labelW, "Background:", val)
			}
		}

		rows = append(rows, rowStyle.Render(prefix+line))
	}

	content := strings.Join(rows, "\n")

	var hint string
	if m.configMsg != "" {
		hint = "\n  " + configHintMutedStyle.Render(m.configMsg)
	}

	box := configBoxBorderStyle.
		Width(w).
		Render("Pipeline Config\n\n" + content + "\n\n  ↑↓ move  Enter re-run  r re-run")

	return box + hint + "\n"
}

func (m tuiModel) handleConfigKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	switch m.configField {
	case cfgFieldSeed:
		switch msg.String() {
		case "up", "down", "tab", "shift+tab", "enter", "r":
		default:
			newInp, cmd := m.seedInput.Update(msg)
			if updated, ok := newInp.(*huh.Input); ok {
				m.seedInput = updated
			}
			return m, cmd
		}

	case cfgFieldBg:
		switch msg.String() {
		case "up", "down", "tab", "shift+tab", "enter", "r":
		default:
			newInp, cmd := m.bgInput.Update(msg)
			if updated, ok := newInp.(*huh.Input); ok {
				m.bgInput = updated
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "up", "k":
		if m.configField > 0 {
			m = m.blurConfigField()
			m.configField--
			m = m.focusConfigField()
			return m.debounceConfigDirty()
		}
	case "down", "j":
		if m.configField < cfgNumFields-1 {
			m = m.blurConfigField()
			m.configField++
			m = m.focusConfigField()
			return m.debounceConfigDirty()
		}

	case "left", "h":
		if m.configField == cfgFieldAlgorithm {
			m.algorithmIdx = (m.algorithmIdx + len(orderedAlgorithms) - 1) % len(orderedAlgorithms)
			return m.debounceConfigDirty()
		}
	case "right", "l":
		if m.configField == cfgFieldAlgorithm {
			m.algorithmIdx = (m.algorithmIdx + 1) % len(orderedAlgorithms)
			return m.debounceConfigDirty()
		}

	case "-":
		switch m.configField {
		case cfgFieldK:
			if m.kVal > 4 {
				m.kVal--
			}
			return m.debounceConfigDirty()
		case cfgFieldIters:
			if m.itersVal > 1 {
				m.itersVal--
			}
			return m.debounceConfigDirty()
		}
	case "=", "+":
		switch m.configField {
		case cfgFieldK:
			if m.kVal < 256 {
				m.kVal++
			}
			return m.debounceConfigDirty()
		case cfgFieldIters:
			if m.itersVal < 100 {
				m.itersVal++
			}
			return m.debounceConfigDirty()
		}

	case " ":
		if m.configField == cfgFieldRetone {
			m.retoneANSI = !m.retoneANSI
			return m.debounceConfigDirty()
		}

	case "o":
		if m.configField == cfgFieldImage {
			m.lastTab = m.tab
			m.state = "picking"
			m.pickPathEdit = false
			m.pickPathBuf = ""
			m.pickPathErr = ""
			return m, m.filePicker.Init()
		}

	case "enter", "r":
		if m.imagePath == "" {
			m.configMsg = "No image loaded. Press o to open file picker."
			return m, nil
		}
		m.configMsg = ""
		m = m.rememberTabForPipelineRun()
		m.state = "loading"
		return m, m.pipelineCmd()
	}

	return m, nil
}
