package cli

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/0mega24/gowall/internal/pipeline"
	"github.com/0mega24/gowall/internal/themes"
)

var templatesPaneBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#555555"))

func (m tuiModel) viewTemplatesTab() string {
	if m.result == nil {
		return "  No palette loaded yet.\n"
	}

	listW := 26
	previewW := m.innerW() - listW - 5
	if previewW < 20 {
		previewW = 20
	}
	boxH := m.height - 10
	if boxH < 4 {
		boxH = 4
	}

	order := m.templateOrder()
	var listLines []string
	for i, id := range order {
		cursor := "  "
		if i == m.templateCursor {
			cursor = "▸ "
		}
		check := "[ ]"
		if m.templates[id] {
			check = "[x]"
		}
		listLines = append(listLines, fmt.Sprintf("%s%s %-14s", cursor, check, id))
	}
	listContent := strings.Join(listLines, "\n")
	leftBox := templatesPaneBorderStyle.
		Width(listW).
		Height(boxH).
		Render(listContent)

	rightView := m.tmplViewport.View()
	if m.contentBuilding[tabTemplates] && m.renderedContent[tabTemplates] == "" {
		rightView = "  building…"
	}
	rightBox := templatesPaneBorderStyle.
		Width(previewW).
		Height(boxH).
		Render(rightView)

	row := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	var statusLine string
	if m.applyErr != "" {
		statusLine = "\n  " + m.applyErr
	}

	return row + statusLine + "\n  Space toggle  ↑↓ navigate  Enter apply  Tab next\n"
}

type templateSnapshot struct {
	result        *pipeline.Result
	templateID    string
	userTemplates []themes.Template
}

func newTemplateSnapshot(m tuiModel) tabBuilder {
	order := m.templateOrder()
	id := ""
	if m.templateCursor >= 0 && m.templateCursor < len(order) {
		id = order[m.templateCursor]
	}
	return templateSnapshot{result: m.result, templateID: id, userTemplates: m.userTemplates}
}

func (s templateSnapshot) build() string {
	if s.result == nil {
		return "No palette loaded yet."
	}
	if bt := themes.BuiltinByID(s.templateID); bt != nil {
		content, err := themes.RenderEmbedded(bt.EmbedPath, s.result.Theme)
		if err != nil {
			return fmt.Sprintf("Error rendering template: %v", err)
		}
		return ColorizeHex(content)
	}
	for _, ut := range s.userTemplates {
		if ut.ID == s.templateID {
			raw, err := os.ReadFile(ut.SourcePath)
			if err != nil {
				return fmt.Sprintf("Error reading template: %v", err)
			}
			content, err := themes.RenderTemplate(string(raw), s.result.Theme)
			if err != nil {
				return fmt.Sprintf("Error rendering template: %v", err)
			}
			return ColorizeHex(content)
		}
	}
	return fmt.Sprintf("Template %q not found.", s.templateID)
}

func (m tuiModel) handleTemplatesKey(msg tea.KeyMsg) (tuiModel, tea.Cmd) {
	order := m.templateOrder()

	switch msg.String() {
	case "up", "k":
		if m.templateCursor > 0 {
			m.templateCursor--
			var mdCmd tea.Cmd
			m, mdCmd = m.markDirty(tabTemplates)
			var vcmd tea.Cmd
			m.tmplViewport, vcmd = m.tmplViewport.Update(msg)
			return m, tea.Batch(mdCmd, vcmd)
		}
	case "down", "j":
		if m.templateCursor < len(order)-1 {
			m.templateCursor++
			var mdCmd tea.Cmd
			m, mdCmd = m.markDirty(tabTemplates)
			var vcmd tea.Cmd
			m.tmplViewport, vcmd = m.tmplViewport.Update(msg)
			return m, tea.Batch(mdCmd, vcmd)
		}
	case " ":
		if m.templateCursor < len(order) {
			id := order[m.templateCursor]
			m.templates[id] = !m.templates[id]
			m.applyErr = ""
		}
	case "enter":
		var ids []string
		for id, on := range m.templates {
			if on {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			m.applyErr = "Select at least one template first."
			return m, nil
		}
		m.applyErr = ""
		m.state = "applying"
		return m, m.applyCmd()
	}

	var cmd tea.Cmd
	m.tmplViewport, cmd = m.tmplViewport.Update(msg)
	return m, cmd
}
