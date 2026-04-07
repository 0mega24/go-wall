package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// composeOverlayAt places overlay string on top of base string at terminal
// position (x, y). It uses ANSI-aware text measurement so escape sequences
// don't corrupt column counts. maxW/maxH are the terminal dimensions.
// Ported from github.com/0mega24/splitr internal/tui.
func composeOverlayAt(base, overlay string, x, y, maxW, maxH int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")
	for oy, oline := range overlayLines {
		by := y + oy
		if by < 0 || by >= len(baseLines) || by >= maxH {
			continue
		}
		if ansi.StringWidth(oline) == 0 {
			continue
		}
		bline := baseLines[by]
		olWidth := ansi.StringWidth(oline)
		prefix := ansi.Cut(bline, 0, x)
		suffix := ansi.Cut(bline, x+olWidth, maxW)
		baseLines[by] = prefix + oline + suffix
	}
	return strings.Join(baseLines, "\n")
}

// tooSmallOverlay renders a centered "terminal too small" box over a blank canvas.
// Uses lipgloss.Place for correct ANSI-aware centering.
func tooSmallOverlay(termW, termH, needW, needH int) string {
	redBold := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	msg := fmt.Sprintf(
		"Terminal size too small\n\nCurrent:   %s × %s\nRequired:  %s × %s\n\nPlease resize your terminal.",
		redBold.Render(fmt.Sprintf("%d", termW)),
		redBold.Render(fmt.Sprintf("%d", termH)),
		dim.Render(fmt.Sprintf("%d", needW)),
		dim.Render(fmt.Sprintf("%d", needH)),
	)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 3).
		Render(msg)

	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, box)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// viewportPlainView renders the visible slice of content for a bubbles viewport
// without calling viewport.Model.View(). The stock View runs lipgloss with
// Width set, which triggers word-wrapping and splits pre-rendered box lines
// (e.g. SectionBox borders) across rows. Scroll state (YOffset, Height,
// SetContent) still comes from the same viewport.Model.
func viewportPlainView(vp viewport.Model, content string) string {
	s := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	h := vp.Height - vp.Style.GetVerticalFrameSize()
	if h < 0 {
		h = 0
	}
	if len(lines) == 0 {
		if h <= 0 {
			return ""
		}
		pad := make([]string, h)
		out := strings.Join(pad, "\n")
		return vp.Style.UnsetWidth().UnsetHeight().Render(out)
	}
	top := max(0, vp.YOffset)
	bottom := clamp(vp.YOffset+h, top, len(lines))
	visible := lines[top:bottom]
	for len(visible) < h {
		visible = append(visible, "")
	}
	if len(visible) > h {
		visible = visible[:h]
	}
	out := strings.Join(visible, "\n")
	return vp.Style.UnsetWidth().UnsetHeight().Render(out)
}

// scrollOverlayConfig controls sizing and the static hint line for a scrollable overlay.
type scrollOverlayConfig struct {
	MinH, MaxH int // box height range inclusive of border; MaxH=0 means use termH
	MinW, MaxW int // box width range inclusive of border; MaxW=0 means use termW
	Hint       string // rendered below box, never scrolls
}

// scrollOverlay is a bordered, scrollable panel with a fixed hint below (splitr-style).
type scrollOverlay struct {
	cfg     scrollOverlayConfig
	vp      viewport.Model
	content string // stored so Resize can re-set after building new viewport
}

func (o *scrollOverlay) SetContent(s string) {
	o.content = s
	o.vp.SetContent(s)
	maxY := max(0, o.vp.TotalLineCount()-o.vp.Height)
	if o.vp.YOffset > maxY {
		o.vp.SetYOffset(maxY)
	}
}

func (o *scrollOverlay) Resize(termW, termH int) {
	_, _, innerW, innerH := o.dimensions(termW, termH)
	prevY := o.vp.YOffset
	vp := viewport.New(innerW, innerH)
	vp.MouseWheelEnabled = true
	vp.SetContent(o.content)
	maxY := max(0, vp.TotalLineCount()-vp.Height)
	vp.SetYOffset(min(prevY, maxY))
	o.vp = vp
}

func (o *scrollOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.vp, cmd = o.vp.Update(msg)
	return cmd
}

// View renders the overlay on base. termW/termH must match Resize (terminal
// size); overlayMaxW/overlayMaxH are the base layer size (e.g. innerW × height).
func (o scrollOverlay) View(base string, termW, termH, overlayMaxW, overlayMaxH int) string {
	boxW, _, _, _ := o.dimensions(termW, termH)
	framed := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(boxW).
		Render(viewportPlainView(o.vp, o.content))

	var stack string
	if o.cfg.Hint != "" {
		hint := lipgloss.NewStyle().Faint(true).Render(o.cfg.Hint)
		stack = lipgloss.JoinVertical(lipgloss.Center, framed, hint)
	} else {
		stack = framed
	}
	sw := lipgloss.Width(stack)
	sh := lipgloss.Height(stack)
	x := max(0, (overlayMaxW-sw)/2)
	y := max(0, (overlayMaxH-sh)/2)
	return composeOverlayAt(base, stack, x, y, overlayMaxW, overlayMaxH)
}

func (o scrollOverlay) dimensions(termW, termH int) (boxW, boxH, innerW, innerH int) {
	maxH := o.cfg.MaxH
	if maxH == 0 {
		maxH = termH
	}
	maxW := o.cfg.MaxW
	if maxW == 0 {
		maxW = termW
	}
	boxW = clamp((termW*4)/5, o.cfg.MinW, maxW)
	boxH = clamp((termH*3)/4, o.cfg.MinH, maxH)
	innerW = max(12, boxW-4)
	innerH = max(4, boxH-4)
	return
}
