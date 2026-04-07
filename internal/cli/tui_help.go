package cli

import "github.com/charmbracelet/lipgloss"

// renderHelpOverlay renders the help content as a centered rounded-border
// box composited over the base terminal content so the background remains visible.
// Sizing and placement use the rendered pane (tab bar + tab body + status), not
// the full terminal height, so short tabs (e.g. Config) do not leave a tall
// empty band below the help box.
func (m tuiModel) renderHelpOverlay(base string) string {
	paneH := lipgloss.Height(base)
	paneW := m.innerW()
	return m.helpOverlay.View(base, m.termW, m.height, paneW, paneH)
}
