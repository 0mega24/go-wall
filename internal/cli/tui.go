package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	icolor "github.com/0mega24/gowall/internal/color"
	"github.com/0mega24/gowall/internal/config"
	"github.com/0mega24/gowall/internal/palette"
	"github.com/0mega24/gowall/internal/pipeline"
	"github.com/0mega24/gowall/internal/themes"
	"github.com/0mega24/gowall/internal/wallpaper"
)

const maxRenderW = 120

// resizeDebounce batches the "all tabs" rebuild after resize settles (inactive tabs).
// The active tab is also refreshed immediately on each WindowSizeMsg (see Update).
const resizeDebounce = 24 * time.Millisecond

// contentDebounce batches config/adjust tab async rebuilds so rapid keypresses do not queue one build per stroke.
const contentDebounce = 30 * time.Millisecond

//go:embed docs/help.md
var helpMarkdown string

const (
	tabConfig    = 0
	tabAdjust    = 1
	tabTemplates = 2
	tabPalette   = 3
	tabPreview   = 4
	numTabs      = 5
)

var orderedAlgorithms = []string{"kmeans++", "kmeans", "mediancut", "octree"}

// ---- messages ---------------------------------------------------------------

type pipelineResult struct {
	result *pipeline.Result
	path   string
	err    error
}

type appliedMsg struct{}

// resizeDebouncedMsg runs after resizeDebounce; seq must match the latest WindowSizeMsg.
type resizeDebouncedMsg struct {
	seq uint64
}

type tabContentMsg struct {
	tab                int
	content            string
	requestedResizeSeq uint64 // m.resizeSeq when the build was scheduled; stale if m.resizeSeq changed since
}

type configDebouncedMsg struct {
	seq uint64
}

type adjustDebouncedMsg struct {
	seq uint64
}

// heartbeatMsg drives a periodic Update tick so spinner animation, pipeline Cmd
// completion, and async tab builds are processed even when stdin is idle (some
// terminals defer the first WindowSizeMsg / wake-ups until focus or first input).
type heartbeatMsg struct{}

// tabBuilder builds cached tab body off the main goroutine.
type tabBuilder interface {
	build() string
}

// ---- model ------------------------------------------------------------------

type tuiModel struct {
	// core state
	imagePath string
	result    *pipeline.Result
	err       error
	state     string // "picking" | "loading" | "preview" | "error" | "applying"
	tab       int    // 0-4
	lastTab   int    // tab to restore after file picking
	quitting  bool

	// wallpaper picking
	filePicker   filepicker.Model
	pickPathEdit bool // true: type a path (Ctrl+p) instead of browsing
	pickPathBuf  string
	pickPathErr  string
	pickColorRef bool // true: pick a Gowall Color Reference file instead of an image

	// pipeline
	spinner    spinner.Model
	customANSI string

	// config tab
	configField  int    // 0=image 1=algorithm 2=seed 3=k 4=iters 5=retone 6=background
	algorithmIdx int    // index into orderedAlgorithms
	seedInputVal string // bound to seedInput
	seedInput    *huh.Input
	kVal         int // cluster count
	itersVal     int // k-means iterations
	retoneANSI   bool
	bgInputVal   string // bound to bgInput
	bgInput      *huh.Input
	configMsg    string // brief status line in config/palette tab

	// adjust tab
	activeSlot     int
	slotMode       [16]string // "auto" | "lock" | "pin"
	lockH          [16]float64
	lockS          [16]float64
	lockV          [16]float64
	hexInputValues [16]string
	hexInputs      [16]*huh.Input
	constraints    map[int]pipeline.SlotConstraint
	globalAdjust   pipeline.GlobalAdjust

	// templates tab
	templates      map[string]bool
	userTemplates  []themes.Template
	templateCursor int
	tmplViewport   viewport.Model
	applyErr       string
	hideBuiltin    bool

	// help overlay
	showHelp    bool
	helpOverlay scrollOverlay

	// palette tab
	swatchMode      int // 0=labeled 1=gradient
	paletteViewport viewport.Model
	// Unverified manual ANSI overrides (palette tab); reapplied after each pipeline run.
	paletteManualANSI     [16]*icolor.Centroid
	paletteManualEditOpen bool
	paletteManualSlot     int
	paletteManualHexBuf   string

	// preview tab
	previewViewport viewport.Model

	// adjust tab pending changes
	adjustPending       bool
	globalAdjustPending bool // true when global H/S/V changed but pipeline not rerun yet
	adjustSection       int  // 0=ANSI, 1=Tones, 2=BG/FG, 3=Global HSV

	// display
	width  int // capped at maxRenderW
	height int
	termW  int // actual terminal width (for centering)

	// resize: increments on each WindowSizeMsg; debounced rebuilds only when seq matches.
	resizeSeq uint64

	// async tab bodies: View reads renderedContent; expensive builds run as tea.Cmd.
	renderedContent   [numTabs]string
	contentDirty      [numTabs]bool
	contentBuilding   [numTabs]bool
	configDebounceSeq uint64
	adjustDebounceSeq uint64

	// lastBuiltResizeSeq records m.resizeSeq when tab body was last applied; used for stale UI hints.
	lastBuiltResizeSeq [numTabs]uint64
}

func newTuiModel(imagePath, customANSI string) tuiModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Spinner{
		Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		FPS:    time.Second / 12,
	}))
	fp := filepicker.New()
	fp.AllowedTypes = []string{".png", ".jpg", ".jpeg", ".webp"}
	fp.AutoHeight = true
	fp.ShowHidden = true
	// Show parent navigation: Back (h/esc/left) moves up; Open (enter/l) enters dirs.
	fp.CurrentDirectory, _ = os.Getwd()
	if fp.CurrentDirectory == "" {
		fp.CurrentDirectory = "."
	}

	m := tuiModel{
		imagePath:  imagePath,
		customANSI: customANSI,
		state:      "loading",
		spinner:    sp,
		filePicker: fp,
		templates:  make(map[string]bool),
		kVal:       32,
		itersVal:   10,
		helpOverlay: scrollOverlay{cfg: scrollOverlayConfig{
			MinH: 10, MinW: 44,
			Hint: "↑↓ jk · pgup/pgdn scroll  ·  ? or esc close",
		}},
	}
	// Load config for default template selection and hide_builtin flag.
	if cfg, err := config.Load(); err == nil {
		for _, id := range cfg.Defaults.Templates {
			m.templates[id] = true
		}
		m.hideBuiltin = cfg.Defaults.HideBuiltin
	}

	// Discover user-defined templates from ~/.config/gowall/templates/
	if ut, err := themes.DiscoverTemplates(); err == nil {
		m.userTemplates = ut
	}

	// Init hex inputs for all 16 adjust slots
	for i := 0; i < 16; i++ {
		m.slotMode[i] = "auto"
		idx := i
		m.hexInputs[idx] = huh.NewInput().
			Placeholder("#rrggbb").
			Value(&m.hexInputValues[idx])
	}

	// Init config tab inputs
	m.seedInputVal = "42"
	m.seedInput = huh.NewInput().
		Placeholder("42").
		Value(&m.seedInputVal)
	m.bgInput = huh.NewInput().
		Placeholder("#000000 or empty").
		Value(&m.bgInputVal)

	return m
}

func heartbeatCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return heartbeatMsg{}
	})
}

func (m tuiModel) Init() tea.Cmd {
	// Run WindowSize before pipeline/picker so m.width/m.termW are set before heavy work.
	// Batch WindowSize with pipeline races the first paint and can leave View stuck until input.
	// Heartbeat keeps the event loop processing timers and async Cmd completions on terminals
	// that barely wake stdin until focus or first click.
	startBeat := tea.Batch(heartbeatCmd(), m.spinner.Tick)
	if m.imagePath != "" {
		return tea.Batch(
			startBeat,
			tea.Sequence(
				tea.WindowSize(),
				func() tea.Msg {
					return runPipeline(m.imagePath, m.retoneANSI, m.customANSI, m.constraints, m.globalAdjust,
						orderedAlgorithms[m.algorithmIdx], parseSeed(m.seedInputVal), m.kVal, m.itersVal, m.bgInputVal)
				},
			),
		)
	}
	path, _, err := wallpaper.FirstOf(wallpaper.DefaultSources()...)
	if err == nil && path != "" {
		m.imagePath = path
		return tea.Batch(
			startBeat,
			tea.Sequence(
				tea.WindowSize(),
				func() tea.Msg {
					return runPipeline(path, m.retoneANSI, m.customANSI, m.constraints, m.globalAdjust,
						orderedAlgorithms[m.algorithmIdx], parseSeed(m.seedInputVal), m.kVal, m.itersVal, m.bgInputVal)
				},
			),
		)
	}
	return tea.Batch(
		startBeat,
		tea.Sequence(
			tea.WindowSize(),
			tea.Batch(
				m.filePicker.Init(),
				func() tea.Msg { return switchStateMsg("picking") },
			),
		),
	)
}

type switchStateMsg string

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case heartbeatMsg:
		if m.quitting {
			return m, nil
		}
		return m, heartbeatCmd()

	case switchStateMsg:
		m.state = string(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.width = msg.Width
		if m.width > maxRenderW {
			m.width = maxRenderW
		}
		m.height = msg.Height
		m.resizeSeq++
		seq := m.resizeSeq
		m.resizeViewports()
		var cmds []tea.Cmd
		// File picker uses AutoHeight; it must receive WindowSizeMsg or Height stays 0 and
		// only one list row is visible.
		var fpCmd tea.Cmd
		m.filePicker, fpCmd = m.filePicker.Update(msg)
		cmds = append(cmds, fpCmd, tea.Tick(resizeDebounce, func(time.Time) tea.Msg {
			return resizeDebouncedMsg{seq: seq}
		}))
		// Refresh visible tab body immediately so layout tracks resize in real time; debounced
		// pass still updates inactive tabs after resize settles.
		if m.showingPreviewUI() {
			var mdCmd tea.Cmd
			m, mdCmd = m.markDirty(m.tab)
			cmds = append(cmds, mdCmd)
		}
		return m, tea.Batch(cmds...)

	case resizeDebouncedMsg:
		if msg.seq != m.resizeSeq {
			return m, nil
		}
		if m.showHelp {
			m = syncHelpOverlayForPane(m)
		}
		if m.result != nil {
			var cmd tea.Cmd
			m, cmd = m.markAllTabsDirty()
			return m, cmd
		}
		return m, nil

	case tabContentMsg:
		if msg.requestedResizeSeq != m.resizeSeq {
			// A newer resize arrived while this build ran; drop and rebuild for current dimensions.
			m.contentBuilding[msg.tab] = false
			var cmd tea.Cmd
			m, cmd = m.markDirty(msg.tab)
			return m, cmd
		}
		m.renderedContent[msg.tab] = msg.content
		m.contentBuilding[msg.tab] = false
		m.lastBuiltResizeSeq[msg.tab] = m.resizeSeq
		switch msg.tab {
		case tabPreview:
			m.previewViewport.SetContent(msg.content)
		case tabPalette:
			m.paletteViewport.SetContent(msg.content)
		case tabTemplates:
			m.tmplViewport.SetContent(msg.content)
		}
		if m.contentDirty[msg.tab] {
			m.contentDirty[msg.tab] = false
			m.contentBuilding[msg.tab] = true
			return m, m.buildTabCmd(msg.tab)
		}
		if m.showHelp {
			m = syncHelpOverlayForPane(m)
		}
		return m, nil

	case configDebouncedMsg:
		if msg.seq != m.configDebounceSeq {
			return m, nil
		}
		return m.markDirty(tabConfig)

	case adjustDebouncedMsg:
		if msg.seq != m.adjustDebounceSeq {
			return m, nil
		}
		return m.markDirty(tabAdjust)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case pipelineResult:
		if msg.err != nil {
			m.err = msg.err
			m.state = "error"
			return m, nil
		}
		m.result = msg.result
		m = m.applyManualANSIOverridesToResult()
		m.imagePath = msg.path
		m.state = "preview"
		m.tab = m.lastTab
		if m.width <= 0 {
			m.width = 80
		}
		if m.height <= 0 {
			m.height = 24
		}
		m.resizeViewports()
		var cmd tea.Cmd
		m, cmd = m.markAllTabsDirty()
		return m, cmd

	case appliedMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}

	// Delegate to sub-models
	if m.state == "picking" {
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)
		return m, cmd
	}

	// Route huh input updates in pin mode (adjust tab)
	if m.showingPreviewUI() && m.tab == tabAdjust {
		slot := m.activeSlot
		if m.adjustSection == 0 && m.slotMode[slot] == "pin" {
			inp := m.hexInputs[slot]
			newInp, cmd := inp.Update(msg)
			if updated, ok := newInp.(*huh.Input); ok {
				m.hexInputs[slot] = updated
			}
			return m, cmd
		}
	}

	// Route huh input updates in config tab (seed/bg fields)
	if m.showingPreviewUI() && m.tab == tabConfig {
		if m.configField == 2 { // seed
			newInp, cmd := m.seedInput.Update(msg)
			if updated, ok := newInp.(*huh.Input); ok {
				m.seedInput = updated
			}
			return m, cmd
		}
		if m.configField == 6 { // background
			newInp, cmd := m.bgInput.Update(msg)
			if updated, ok := newInp.(*huh.Input); ok {
				m.bgInput = updated
			}
			return m, cmd
		}
	}

	// Route viewport updates
	if m.showingPreviewUI() {
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
		}
	}

	// Route help overlay when open
	if m.showHelp {
		return m, m.helpOverlay.Update(msg)
	}

	return m, nil
}

func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay intercepts everything except close keys
	if m.showHelp {
		switch msg.String() {
		case "?", "esc", "q":
			m.showHelp = false
			m.helpOverlay.cfg.MaxH = 0
			m.helpOverlay.cfg.MaxW = 0
			return m, nil
		default:
			return m, m.helpOverlay.Update(msg)
		}
	}

	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "q":
		if m.state != "preview" {
			m.quitting = true
			return m, tea.Quit
		}
		m.quitting = true
		return m, tea.Quit
	case "?":
		if m.showingPreviewUI() {
			m.showHelp = true
			m = syncHelpOverlayForPane(m)
		}
		return m, nil
	case "tab":
		if m.showingPreviewUI() {
			m.tab = (m.tab + 1) % numTabs
		}
		return syncHelpOverlayForPane(m), nil
	case "shift+tab":
		if m.showingPreviewUI() {
			m.tab = (m.tab + numTabs - 1) % numTabs
		}
		return syncHelpOverlayForPane(m), nil
	case "o":
		if m.showingPreviewUI() {
			m.lastTab = m.tab
			m.pickColorRef = false
			m.filePicker.AllowedTypes = []string{".png", ".jpg", ".jpeg", ".webp"}
			m.state = "picking"
			m.pickPathEdit = false
			m.pickPathBuf = ""
			m.pickPathErr = ""
			return m, m.filePicker.Init()
		}
	case "I":
		if m.showingPreviewUI() {
			m.lastTab = m.tab
			m.pickColorRef = true
			m.filePicker.AllowedTypes = nil // any file
			m.state = "picking"
			m.pickPathEdit = false
			m.pickPathBuf = ""
			m.pickPathErr = ""
			return m, m.filePicker.Init()
		}
	case "ctrl+p":
		if m.state == "picking" {
			m.pickPathEdit = !m.pickPathEdit
			m.pickPathErr = ""
			if m.pickPathEdit {
				m.pickPathBuf = m.filePicker.CurrentDirectory
			}
			return m, nil
		}
	}

	switch m.state {
	case "picking":
		if m.pickPathEdit {
			return m.handlePickPathKey(msg)
		}
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)
		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			if m.pickColorRef {
				return m.finishColorReferenceImport(path)
			}
			m.imagePath = path
			m = m.rememberTabForPipelineRun()
			m.state = "loading"
			m.pickPathEdit = false
			m.pickPathBuf = ""
			m.pickPathErr = ""
			return m, tea.Batch(
				m.spinner.Tick,
				func() tea.Msg {
					return runPipeline(path, m.retoneANSI, m.customANSI, m.constraints, m.globalAdjust,
						orderedAlgorithms[m.algorithmIdx], parseSeed(m.seedInputVal), m.kVal, m.itersVal, m.bgInputVal)
				},
			)
		}
		return m, cmd
	case "preview", "loading":
		if m.state == "loading" && m.result == nil {
			return m, nil
		}
		return m.handlePreviewKey(msg)
	}
	return m, nil
}

// finishColorReferenceImport loads a palette export / color reference file and
// replaces the current result without running the image pipeline.
func (m tuiModel) finishColorReferenceImport(path string) (tea.Model, tea.Cmd) {
	theme, err := themes.LoadColorReferenceFile(path)
	if err != nil {
		m.pickPathErr = err.Error()
		return m, nil
	}
	res, err := pipeline.ResultFromThemeData(theme)
	if err != nil {
		m.pickPathErr = err.Error()
		return m, nil
	}
	m.result = &res
	m = m.applyManualANSIOverridesToResult()
	m.imagePath = path
	m.pickColorRef = false
	m.filePicker.AllowedTypes = []string{".png", ".jpg", ".jpeg", ".webp"}
	m.state = "preview"
	m.pickPathEdit = false
	m.pickPathBuf = ""
	m.pickPathErr = ""
	m.tab = m.lastTab
	if m.width <= 0 {
		m.width = 80
	}
	if m.height <= 0 {
		m.height = 24
	}
	m.resizeViewports()
	var cmd tea.Cmd
	m, cmd = m.markAllTabsDirty()
	return m, cmd
}

func expandUserPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return filepath.Abs(path)
}

func (m tuiModel) applyPickPath() (tea.Model, tea.Cmd) {
	m.pickPathErr = ""
	raw := strings.TrimSpace(m.pickPathBuf)
	if raw == "" {
		m.pickPathErr = "empty path"
		return m, nil
	}
	abs, err := expandUserPath(raw)
	if err != nil {
		m.pickPathErr = err.Error()
		return m, nil
	}
	st, err := os.Stat(abs)
	if err != nil {
		m.pickPathErr = err.Error()
		return m, nil
	}
	if st.IsDir() {
		m.filePicker.CurrentDirectory = abs
		m.pickPathEdit = false
		m.pickPathBuf = ""
		return m, m.filePicker.Init()
	}
	if m.pickColorRef {
		return m.finishColorReferenceImport(abs)
	}
	ext := filepath.Ext(abs)
	ok := false
	for _, e := range m.filePicker.AllowedTypes {
		if strings.EqualFold(ext, e) {
			ok = true
			break
		}
	}
	if !ok {
		m.pickPathErr = "not an image file (.png .jpg .jpeg .webp)"
		return m, nil
	}
	path := abs
	m.imagePath = path
	m = m.rememberTabForPipelineRun()
	m.state = "loading"
	m.pickPathEdit = false
	m.pickPathBuf = ""
	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return runPipeline(path, m.retoneANSI, m.customANSI, m.constraints, m.globalAdjust,
				orderedAlgorithms[m.algorithmIdx], parseSeed(m.seedInputVal), m.kVal, m.itersVal, m.bgInputVal)
		},
	)
}

func (m tuiModel) handlePickPathKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type { //nolint:exhaustive
	case tea.KeyEnter:
		return m.applyPickPath()
	case tea.KeyEscape:
		m.pickPathEdit = false
		m.pickPathErr = ""
		return m, nil
	case tea.KeyBackspace, tea.KeyCtrlH:
		if len(m.pickPathBuf) > 0 {
			r := []rune(m.pickPathBuf)
			m.pickPathBuf = string(r[:len(r)-1])
		}
		m.pickPathErr = ""
		return m, nil
	default:
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			m.pickPathBuf += string(msg.Runes)
			m.pickPathErr = ""
		}
		return m, nil
	}
}

func (m tuiModel) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.tab {
	case tabConfig:
		nm, cmd := m.handleConfigKey(msg)
		return nm, cmd
	case tabAdjust:
		nm, cmd := m.handleAdjustKey(msg)
		return nm, cmd
	case tabTemplates:
		nm, cmd := m.handleTemplatesKey(msg)
		return nm, cmd
	case tabPalette:
		nm, cmd := m.handlePaletteKey(msg)
		return nm, cmd
	case tabPreview:
		var cmd tea.Cmd
		m.previewViewport, cmd = m.previewViewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

// innerW returns the available inner content width (terminal width minus outer border).
func (m tuiModel) innerW() int {
	w := m.width - 2
	if w < 20 {
		return 20
	}
	return w
}

// showingPreviewUI is true when the tabbed main UI is shown and should receive
// keys and viewport updates (including pipeline re-runs while a result is kept visible).
func (m tuiModel) showingPreviewUI() bool {
	if m.state == "preview" {
		return true
	}
	return m.state == "loading" && m.result != nil
}

func (m tuiModel) markDirty(tabs ...int) (tuiModel, tea.Cmd) {
	var cmds []tea.Cmd
	for _, tab := range tabs {
		if tab < 0 || tab >= numTabs {
			continue
		}
		m.contentDirty[tab] = true
		if !m.contentBuilding[tab] {
			m.contentBuilding[tab] = true
			m.contentDirty[tab] = false // this build satisfies the dirty request
			cmds = append(cmds, m.buildTabCmd(tab))
		}
	}
	return m, tea.Batch(cmds...)
}

func (m tuiModel) markAllTabsDirty() (tuiModel, tea.Cmd) {
	return m.markDirty(tabConfig, tabAdjust, tabTemplates, tabPalette, tabPreview)
}

func (m tuiModel) buildTabCmd(tab int) tea.Cmd {
	rs := m.resizeSeq
	snap := m.snapshotForTab(tab)
	return func() tea.Msg {
		return tabContentMsg{tab: tab, content: snap.build(), requestedResizeSeq: rs}
	}
}

func (m tuiModel) debounceConfigDirty() (tuiModel, tea.Cmd) {
	m.configDebounceSeq++
	seq := m.configDebounceSeq
	return m, tea.Tick(contentDebounce, func(time.Time) tea.Msg {
		return configDebouncedMsg{seq: seq}
	})
}

func (m tuiModel) debounceAdjustDirty() (tuiModel, tea.Cmd) {
	m.adjustDebounceSeq++
	seq := m.adjustDebounceSeq
	return m, tea.Tick(contentDebounce, func(time.Time) tea.Msg {
		return adjustDebouncedMsg{seq: seq}
	})
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}
	// Do not short-circuit on m.width < 4: innerW() already clamps when width is still 0
	// before the first WindowSizeMsg, and an early return here prevented renderBase() from
	// running until a click woke the terminal (pipeline could finish but View stayed stuck).
	// Too-small check uses actual terminal width.
	if m.termW > 0 && m.termW < 80 {
		return tooSmallOverlay(m.termW, m.height, 80, 24)
	}

	base := m.renderBase()
	if m.showHelp {
		base = m.renderHelpOverlay(base)
	}

	// Wrap in outer border container.
	box := outerBorderStyle.Width(m.innerW()).Render(base)

	// Center horizontally if terminal is wider than our render width.
	if m.termW > m.width {
		return lipgloss.PlaceHorizontal(m.termW, lipgloss.Center, box)
	}
	return box
}

func (m tuiModel) renderBase() string {
	switch m.state {
	case "picking":
		return m.viewPicking()
	case "loading":
		if m.result == nil {
			return m.viewLoading()
		}
		return m.viewPreviewContent()
	case "error":
		return m.viewError()
	case "applying":
		return "  Applying…\n"
	case "preview":
		return m.viewPreviewContent()
	}
	return "\n  Loading…\n"
}

func (m tuiModel) viewPreviewContent() string {
	var b strings.Builder
	b.WriteString(m.viewTabBar())

	switch m.tab {
	case tabConfig:
		b.WriteString(m.viewConfigTab())
	case tabAdjust:
		b.WriteString(m.viewAdjustTab())
	case tabTemplates:
		b.WriteString(m.viewTemplatesTab())
	case tabPalette:
		b.WriteString(m.viewPaletteTab())
	case tabPreview:
		b.WriteString(m.viewPreviewTab())
	}

	b.WriteString(m.viewStatusBar())
	return b.String()
}

func (m tuiModel) viewPicking() string {
	var b strings.Builder
	if m.pickColorRef {
		b.WriteString("\n  Import Gowall Color Reference (palette export format)\n\n")
	} else {
		b.WriteString("\n  Select a wallpaper image\n\n")
	}
	fmt.Fprintf(&b, "  %s\n\n", m.filePicker.CurrentDirectory)
	b.WriteString(m.filePicker.View())
	if m.pickPathErr != "" {
		fmt.Fprintf(&b, "\n  %s\n", m.pickPathErr)
	}
	if m.pickPathEdit {
		b.WriteString("  Path: ")
		b.WriteString(m.pickPathBuf)
		b.WriteString("\n")
		b.WriteString("\n  enter: go  esc: cancel  ctrl+p: close path  q quit\n")
	} else {
		b.WriteString("\n  j/k move  l/enter open folder or file  h parent dir  ctrl+p: type path  q quit\n")
	}
	b.WriteString(m.viewStatusBar())
	return b.String()
}

func (m tuiModel) viewLoading() string {
	var b strings.Builder
	b.WriteString(m.viewTabBar())
	fmt.Fprintf(&b, "\n\n  %s  Extracting colors from image…\n", m.spinner.View())
	b.WriteString(m.viewStatusBar())
	return b.String()
}

func (m tuiModel) viewError() string {
	return fmt.Sprintf("\n  Error: %v\n\n  q quit\n", m.err)
}

// tabBorderOpen returns a lipgloss border for the active (open-bottom) folder tab.
func tabBorderOpen() lipgloss.Border {
	b := lipgloss.RoundedBorder()
	b.BottomLeft = "╯"
	b.Bottom = " "
	b.BottomRight = "╰"
	return b
}

// tabBorderClosed returns a lipgloss border for inactive folder tabs.
func tabBorderClosed() lipgloss.Border {
	b := lipgloss.RoundedBorder()
	b.BottomLeft = "┴"
	b.BottomRight = "┴"
	return b
}

// Package-level styles (fixed colors/borders) — avoid per-frame allocation.
var (
	activeTabStyle = lipgloss.NewStyle().
			Border(tabBorderOpen(), true).
			Padding(0, 1).
			Bold(true)

	inactiveTabStyle = lipgloss.NewStyle().
				Border(tabBorderClosed(), true).
				Padding(0, 1).
				Faint(true)

	tabHintFaintStyle = lipgloss.NewStyle().Faint(true)

	statusBarBaseStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1e1e1e")).
				Foreground(lipgloss.Color("#777777"))

	outerBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444"))
)

func (m tuiModel) viewTabBar() string {
	names := []string{"Config", "Adjust", "Templates", "Palette", "Preview"}

	hintText := "? help  o open  I import  q quit  "
	hint := tabHintFaintStyle.Render("  " + hintText)
	hintW := lipgloss.Width(hint)

	var topB, midB, botB strings.Builder
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
		top, mid, bot := lines[0], lines[1], lines[2]
		if i > 0 {
			topB.WriteByte(' ')
			midB.WriteByte(' ')
			botB.WriteString("─")
			totalTabsWidth++
		}
		topB.WriteString(top)
		midB.WriteString(mid)
		botB.WriteString(bot)
		totalTabsWidth += lipgloss.Width(top)
	}

	innerW := m.innerW()
	// If tabs + hint do not fit on one row, put the hint on its own line below the
	// tab row so the terminal does not wrap mid-line and break the border art.
	if totalTabsWidth+hintW > innerW {
		fillW := innerW - totalTabsWidth
		if fillW < 0 {
			fillW = 0
		}
		top := topB.String() + strings.Repeat(" ", fillW)
		mid := midB.String() + strings.Repeat(" ", fillW)
		bot := botB.String() + strings.Repeat("─", fillW)
		hintLine := ansi.Cut(hint, 0, innerW)
		return top + "\n" + mid + "\n" + bot + "\n" + hintLine + "\n"
	}

	fillW := innerW - totalTabsWidth - hintW
	if fillW < 0 {
		fillW = 0
	}

	top := topB.String() + strings.Repeat(" ", fillW+hintW)
	mid := midB.String() + strings.Repeat(" ", fillW) + hint
	bot := botB.String() + strings.Repeat("─", fillW+hintW)

	return top + "\n" + mid + "\n" + bot + "\n"
}

// viewStatusBar renders a 1-line status bar showing context about the current state.
func (m tuiModel) viewStatusBar() string {
	imgName := "(no image)"
	if m.imagePath != "" {
		parts := strings.Split(m.imagePath, "/")
		imgName = parts[len(parts)-1]
	}
	algo := orderedAlgorithms[m.algorithmIdx]
	tabNames := []string{"Config", "Adjust", "Templates", "Palette", "Preview"}
	tabLabel := ""
	if m.tab >= 0 && m.tab < len(tabNames) {
		tabLabel = "[" + tabNames[m.tab] + "]"
	}
	left := fmt.Sprintf("  %s · %s · k=%d", imgName, algo, m.kVal)
	if m.hasManualANSIOverrides() {
		left += " · unverified ANSI"
	}
	right := tabLabel + "  "
	if m.state == "loading" && m.result != nil {
		right = m.spinner.View() + " running…  "
	} else {
		switch m.tab {
		case tabPalette:
			right = tabLabel + "  ↑↓ scroll · m manual · u clear · Tab · q  "
		case tabPreview:
			right = tabLabel + "  ↑↓ PgUp/Dn · Tab · q  "
		}
	}

	leftStr := statusBarBaseStyle.Render(left)
	rightStr := statusBarBaseStyle.Render(right)
	fillW := m.innerW() - lipgloss.Width(leftStr) - lipgloss.Width(rightStr)
	if fillW < 0 {
		fillW = 0
	}
	fill := statusBarBaseStyle.Width(fillW).Render("")
	return leftStr + fill + rightStr
}

func (m tuiModel) hasManualANSIOverrides() bool {
	for _, p := range m.paletteManualANSI {
		if p != nil {
			return true
		}
	}
	return false
}

// applyManualANSIOverridesToResult patches Theme.Ansi / result.ANSI from palette tab overrides.
func (m tuiModel) applyManualANSIOverridesToResult() tuiModel {
	if m.result == nil {
		return m
	}
	for i := 0; i < 16; i++ {
		if m.paletteManualANSI[i] != nil {
			m.result.ANSI[i] = *m.paletteManualANSI[i]
			m.result.Theme.Ansi[i] = m.paletteManualANSI[i].RawHex()
		}
	}
	return m
}

// ---- pipeline ---------------------------------------------------------------

func runPipeline(path string, retone bool, customANSI string,
	constraints map[int]pipeline.SlotConstraint,
	globalAdjust pipeline.GlobalAdjust,
	algo string, seed int64, kVal, itersVal int, bgHex string,
) tea.Msg {
	img, err := wallpaper.LoadImage(path)
	if err != nil {
		return pipelineResult{err: err}
	}
	opts := pipeline.DefaultOptions()
	opts.RetoneANSI = retone
	opts.Seed = seed
	opts.Constraints = constraints
	opts.GlobalAdjust = globalAdjust
	if kVal > 0 {
		opts.KMeansK = kVal
	}
	if itersVal > 0 {
		opts.KMeansIters = itersVal
	}
	if bgHex != "" {
		opts.BackgroundHex = bgHex
	}
	if algo != "" && algo != "kmeans++" {
		c, err := icolor.Get(algo)
		if err == nil {
			opts.Clusterer = c
		}
	}
	if customANSI != "" {
		custom, err := palette.LoadANSIHexFile(customANSI)
		if err != nil {
			return pipelineResult{err: err}
		}
		opts.CustomANSI = custom
		opts.RetoneANSI = true
	}
	result, err := pipeline.Run(img, opts, nil)
	if err != nil {
		return pipelineResult{err: err}
	}
	return pipelineResult{result: &result, path: path}
}

// rememberTabForPipelineRun records the active tab so pipelineResult can restore
// it after a rerun. Without this, lastTab stays at its default (Config) or a
// stale value from the file picker, and saving from Adjust jumps back to Config.
func (m tuiModel) rememberTabForPipelineRun() tuiModel {
	m.lastTab = m.tab
	return m
}

// pipelineCmd returns a tea.Cmd that runs the pipeline.
// Callers must set m.state = "loading" before returning.
func (m tuiModel) pipelineCmd() tea.Cmd {
	algo := orderedAlgorithms[m.algorithmIdx]
	seed := parseSeed(m.seedInputVal)
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return runPipeline(m.imagePath, m.retoneANSI, m.customANSI, m.constraints, m.globalAdjust,
				algo, seed, m.kVal, m.itersVal, m.bgInputVal)
		},
	)
}

// parseSeed parses a seed string, returning def on failure.
func parseSeed(s string) int64 {
	if v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
		return v
	}
	return 42
}

// ---- apply ------------------------------------------------------------------

func (m tuiModel) applyCmd() tea.Cmd {
	return func() tea.Msg {
		var ids []string
		for id, on := range m.templates {
			if on {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 || m.result == nil {
			return appliedMsg{}
		}
		_ = ApplyTheme(m.result.Theme, ids)
		return appliedMsg{}
	}
}

func (m tuiModel) templateOrder() []string {
	order := make([]string, 0, len(themes.BuiltinTemplates)+len(m.userTemplates))
	if !m.hideBuiltin {
		for _, t := range themes.BuiltinTemplates {
			order = append(order, t.ID)
		}
	}
	for _, t := range m.userTemplates {
		if !t.IsBuiltin {
			order = append(order, t.ID)
		}
	}
	return order
}

// ---- help rendering ---------------------------------------------------------

var (
	helpRendererMu sync.Mutex
	helpRenderers  = make(map[int]*glamour.TermRenderer)
)

const maxHelpRendererCache = 32

func rendererForWidth(width int) (*glamour.TermRenderer, error) {
	helpRendererMu.Lock()
	defer helpRendererMu.Unlock()
	if r, ok := helpRenderers[width]; ok {
		return r, nil
	}
	if len(helpRenderers) >= maxHelpRendererCache {
		for k := range helpRenderers {
			delete(helpRenderers, k)
			break
		}
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}
	helpRenderers[width] = r
	return r, nil
}

func renderHelp(width int) string {
	if width < 20 {
		width = 80
	}
	r, err := rendererForWidth(width)
	if err != nil {
		return helpMarkdown
	}
	out, err := r.Render(helpMarkdown)
	if err != nil {
		return helpMarkdown
	}
	return out
}

// ---- viewport helpers -------------------------------------------------------

// syncHelpOverlayForPane sizes the help overlay to the current rendered preview
// pane (tab bar + body + status) so the box height tracks short vs tall tabs.
func syncHelpOverlayForPane(m tuiModel) tuiModel {
	if !m.showHelp {
		m.helpOverlay.cfg.MaxH = 0
		m.helpOverlay.cfg.MaxW = 0
		m.helpOverlay.Resize(m.termW, m.height)
		return m
	}
	paneH := lipgloss.Height(m.renderBase())
	m.helpOverlay.cfg.MaxH = paneH
	m.helpOverlay.cfg.MaxW = m.innerW()
	_, _, iw, _ := m.helpOverlay.dimensions(m.termW, m.height)
	m.helpOverlay.SetContent(renderHelp(iw))
	m.helpOverlay.Resize(m.termW, m.height)
	return m
}

// resizeViewports recalculates all viewport dimensions from m.width/m.height.
// Heights account for: outer border (-2), tab bar (-4), status bar (-1).
func (m *tuiModel) resizeViewports() {
	iw := m.innerW()
	listW := 26

	// Template viewport: inner box border adds 2 more.
	vpH := m.height - 13
	if vpH < 4 {
		vpH = 4
	}
	paletteH := m.height - 11
	if paletteH < 4 {
		paletteH = 4
	}
	// Preview: subtract 1 extra for the loading bar that may appear.
	previewH := m.height - 13
	if previewH < 4 {
		previewH = 4
	}

	m.tmplViewport = viewport.New(iw-listW-7, vpH)
	*m = syncHelpOverlayForPane(*m)
	// Use full inner width so SectionBox (boxW = vpWidth-4, outer = vpWidth-4+4 = iw)
	// fills the outer container exactly, matching how the Config tab box renders.
	m.previewViewport = viewport.New(iw, previewH)
	m.paletteViewport = viewport.New(iw, paletteH)
}

// emptySnapshot is a fallback for unknown tab indices.
type emptySnapshot struct{}

func (emptySnapshot) build() string { return "" }

func (m tuiModel) snapshotForTab(tab int) tabBuilder {
	switch tab {
	case tabConfig:
		return newConfigSnapshot(m)
	case tabAdjust:
		return newAdjustSnapshot(m)
	case tabTemplates:
		return newTemplateSnapshot(m)
	case tabPalette:
		return newPaletteSnapshot(m)
	case tabPreview:
		return newPreviewSnapshot(m)
	default:
		return emptySnapshot{}
	}
}

// ---- RunTUI -----------------------------------------------------------------

// RunTUI starts the interactive TUI.
func RunTUI(imagePath string) error {
	m := newTuiModel(imagePath, "")
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// RunTUIConfig starts the TUI with full config options.
func RunTUIConfig(imagePath, customANSI string) error {
	m := newTuiModel(imagePath, customANSI)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
