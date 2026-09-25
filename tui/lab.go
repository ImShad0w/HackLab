package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hacklab/internal/lab"
	"hacklab/internal/progress"
	"hacklab/internal/session"
)

// Colors
const (
	accentColor = "#e94560"
	cyanColor   = "#00d4ff"
	greenColor  = "#00ff88"
	yellowColor = "#f0e68c"
	dimColor    = "#444466"
	borderColor = "#1a1a3e"
)

// HACKLAB ASCII logo — same as the CLI banner
const asciiLogo = `HACKLAB`

// Styles
var (
	logoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor)).Bold(true)
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor)).Bold(true)
	taglineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8888aa")).Italic(true)
	urlStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor)).Underline(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(yellowColor)).Bold(true)
	tagStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor))
	progressMsg  = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor))
	sepStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))
	footerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(dimColor))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(yellowColor)).Bold(true)

	btnBase = lipgloss.NewStyle().
		Background(lipgloss.Color("#23234a")).
		Foreground(lipgloss.Color("#9a9ab5")).
		Padding(0, 4).
		Bold(true)
	btnSelected = lipgloss.NewStyle().
			Background(lipgloss.Color(accentColor)).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 4).
			Bold(true)

	objNameStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0")).Bold(true)
	objDoneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor)).Bold(true)
	objSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor)).Bold(true)
	categoryStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(dimColor)).Padding(0, 1).MarginLeft(1)
	hintStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(yellowColor)).PaddingLeft(4).Italic(true)
	progressBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor))
	progressDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))
	arrowStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor)).Bold(true)

	checkDone     = lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor)).Render("✔")
	checkEmpty    = lipgloss.NewStyle().Foreground(lipgloss.Color(dimColor)).Render("○")
	checkSelected = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor)).Render("◉")
)

// Phase represents the current view state
type Phase int

const (
	PhaseWelcome Phase = iota
	PhaseQuiz
	PhaseComplete
	PhaseWipeConfirm
)

type model struct {
	lab       *lab.Lab
	prog      *progress.Progress
	phase     Phase
	lastPhase Phase // phase to return to when a modal (wipe confirm) is cancelled
	cursor    int
	scroll    int
	width     int
	height    int
	targetURL string
	showHints map[int]bool
	wipeIdx   int  // 0 = cancel, 1 = accept (wipe session)
	wiped     bool // true once the user accepted wiping the session
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		switch m.phase {
		case PhaseWelcome:
			switch msg.String() {
			case "enter", " ":
				m.phase = PhaseQuiz
				m.cursor = 0
				m.scroll = 0
				m.showHints = make(map[int]bool)
			case "x":
				m.openWipeConfirm()
			}

		case PhaseQuiz:
			total := len(m.lab.Manifest.Objectives)
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					if m.cursor < m.scroll {
						m.scroll = m.cursor
					}
				}
			case "down", "j":
				if m.cursor < total-1 {
					m.cursor++
					visibleArea := m.getVisibleArea()
					if m.cursor >= m.scroll+visibleArea {
						m.scroll = m.cursor - visibleArea + 1
					}
				}
			case " ", "enter":
				m.toggleObjective(m.cursor)
			case "h", "H":
				m.showHints[m.cursor] = !m.showHints[m.cursor]
			case "x":
				m.openWipeConfirm()
			case "q":
				return m, tea.Quit
			}

		case PhaseWipeConfirm:
			switch msg.String() {
			case "left", "h", "up", "k":
				m.wipeIdx = 0
			case "right", "l", "down", "j":
				m.wipeIdx = 1
			case "enter", " ":
				if m.wipeIdx == 1 {
					m.wipe()
					return m, tea.Quit
				}
				m.phase = m.lastPhase
			case "esc", "q":
				m.phase = m.lastPhase
			}

		case PhaseComplete:
			if msg.String() == "enter" || msg.String() == " " || msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// openWipeConfirm shows the wipe-session confirmation dialog.
func (m *model) openWipeConfirm() {
	m.lastPhase = m.phase
	m.wipeIdx = 0 // default to cancel — destructive action needs an explicit choice
	m.phase = PhaseWipeConfirm
}

// wipe erases the lab's progress and the saved session, so the session
// can no longer be resumed. It must always be reached via the confirmation
// dialog — this is destructive and cannot be undone.
func (m *model) wipe() {
	m.prog.WipeLab(m.lab.Name)
	_ = m.prog.Save()
	_ = session.Clear()
	m.wiped = true
}

func (m *model) toggleObjective(idx int) {
	wasCompleted := m.prog.IsCompleted(m.lab.Name, idx)
	if wasCompleted {
		var newCompleted []int
		for _, i := range m.prog.Labs[m.lab.Name].Completed {
			if i != idx {
				newCompleted = append(newCompleted, i)
			}
		}
		m.prog.Labs[m.lab.Name].Completed = newCompleted
	} else {
		m.prog.CompleteObjective(m.lab.Name, idx)
	}
	_ = m.prog.Save()

	completed, _ := m.prog.LabStats(m.lab.Name)
	if completed == len(m.lab.Manifest.Objectives) {
		m.phase = PhaseComplete
	}
}

func (m model) getVisibleArea() int {
	used := 10
	return m.height - used
}

func (m model) View() string {
	switch m.phase {
	case PhaseWelcome:
		return m.viewWelcome()
	case PhaseQuiz:
		return m.viewQuiz()
	case PhaseComplete:
		return m.viewComplete()
	case PhaseWipeConfirm:
		return m.viewWipeConfirm()
	default:
		return ""
	}
}

// viewWipeConfirm renders the destructive-action confirmation dialog.
// The user must explicitly move to Accept and hit enter — progress is lost
// and the session can no longer be resumed.
func (m model) viewWipeConfirm() string {
	w := m.width
	h := m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	content := strings.Join([]string{
		"🧹 " + titleStyle.Render("Wipe session"),
		"",
		taglineStyle.Render("Lab: " + m.lab.Manifest.Name),
		"",
		"  This will permanently erase all progress for this lab.",
		"  The session will be closed and can no longer be resumed.",
		"",
		warnStyle.Render("⚠  Your progress will be lost. This cannot be undone."),
		"",
		strings.Join([]string{
			m.wipeButton("Cancel", m.wipeIdx == 0),
			m.wipeButton("Accept", m.wipeIdx == 1),
		}, "     "),
		"",
		footerStyle.Render("←/→ or j/k to choose  ·  enter to confirm  ·  esc to go back"),
	}, "\n")

	boxW := min(w-4, 72)
	if boxW < 40 {
		boxW = 40
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(accentColor)).
		Padding(1, 3).
		Width(boxW).
		Align(lipgloss.Center)

	rendered := box.Render(content)

	// Center the box both horizontally and vertically in the terminal.
	padLeft := (w - boxW) / 2
	if padLeft < 0 {
		padLeft = 0
	}

	lines := strings.Split(rendered, "\n")
	padTop := 0
	if h > len(lines) {
		padTop = (h - len(lines)) / 2
	}

	var b strings.Builder
	b.WriteString(strings.Repeat("\n", padTop))
	for _, ln := range lines {
		b.WriteString(strings.Repeat(" ", padLeft))
		b.WriteString(ln)
		b.WriteString("\n")
	}
	return b.String()
}

func (m model) wipeButton(label string, selected bool) string {
	if selected {
		return btnSelected.Render(label)
	}
	return btnBase.Render(label)
}

func (m model) viewWelcome() string {
	mf := m.lab.Manifest
	w := m.width
	if w <= 0 {
		w = 80
	}

	lines := m.buildWelcomeLines(mf, w)

	// Vertical centering
	totalLines := len(lines)
	padTop := 0
	if m.height > totalLines {
		padTop = (m.height - totalLines) / 2
	}

	var b strings.Builder
	for i := 0; i < padTop; i++ {
		b.WriteString("\n")
	}
	for _, line := range lines {
		b.WriteString(line + "\n")
	}

	return b.String()
}

func (m model) buildWelcomeLines(mf *lab.Manifest, w int) []string {
	// Styles with built-in centering (handles ANSI correctly)
	logoLine := logoStyle.Copy().Align(lipgloss.Center).Width(w)
	tagLine := taglineStyle.Copy().Align(lipgloss.Center).Width(w)
	titleLine := titleStyle.Copy().Align(lipgloss.Center).Width(w)
	infoLine := infoStyle.Copy().Align(lipgloss.Center).Width(w)
	tagLineCentered := tagStyle.Copy().Align(lipgloss.Center).Width(w)
	progLine := progressMsg.Copy().Align(lipgloss.Center).Width(w)
	footerLine := footerStyle.Copy().Align(lipgloss.Center).Width(w)
	sepLine := sepStyle.Copy().Align(lipgloss.Center).Width(w)

	sepStr := strings.Repeat("─", min(w, 60))

	var lines []string

	// ASCII logo — green, bold, centered by lipgloss
	for _, ll := range strings.Split(asciiLogo, "\n") {
		if ll == "" {
			continue
		}
		lines = append(lines, logoLine.Render(ll))
	}

	lines = append(lines, "")
	lines = append(lines, tagLine.Render("your terminal hacking playground"))
	lines = append(lines, "")
	lines = append(lines, sepLine.Render(sepStr))
	lines = append(lines, "")

	// Lab name
	lines = append(lines, titleLine.Render(mf.Name))

	// Description
	if mf.Description != "" {
		lines = append(lines, tagLine.Render(mf.Description))
	}

	lines = append(lines, "")

	// Difficulty + objectives
	difficulty := "UNKNOWN"
	if mf.Difficulty != "" {
		difficulty = strings.ToUpper(mf.Difficulty)
	}
	info := fmt.Sprintf("Difficulty: %s  ·  Objectives: %d", difficulty, len(mf.Objectives))
	lines = append(lines, infoLine.Render(info))

	// Tags
	if len(mf.Tags) > 0 {
		lines = append(lines, "")
		lines = append(lines, tagLineCentered.Render(strings.Join(mf.Tags, "  ")))
	}

	// Previous progress
	completed, _ := m.prog.LabStats(m.lab.Name)
	if completed > 0 {
		lines = append(lines, "")
		lines = append(lines, progLine.Render(fmt.Sprintf("Previously completed: %d/%d", completed, len(mf.Objectives))))
	}

	lines = append(lines, "")
	lines = append(lines, sepLine.Render(sepStr))
	lines = append(lines, "")
	lines = append(lines, footerLine.Render("press enter to begin  ·  x wipe session  ·  q to quit"))

	return lines
}

func (m model) viewQuiz() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	mf := m.lab.Manifest
	total := len(mf.Objectives)
	completed, _ := m.prog.LabStats(m.lab.Name)
	pct := 0.0
	if total > 0 {
		pct = float64(completed) / float64(total) * 100
	}

	var b strings.Builder

	// === HEADER ===
	b.WriteString(titleStyle.Render(" ⚡ "+mf.Name) + " ")
	b.WriteString(footerStyle.Render(mf.Difficulty) + "\n")

	if m.targetURL != "" {
		b.WriteString(" 📡 " + urlStyle.Render(m.targetURL) + "\n")
	}
	b.WriteString("\n")

	// Progress bar
	barWidth := w - 20
	if barWidth < 20 {
		barWidth = 20
	}
	filled := int(pct / 100 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := progressBarStyle.Render(strings.Repeat("█", filled))
	bar += progressDimStyle.Render(strings.Repeat("░", barWidth-filled))
	b.WriteString(fmt.Sprintf("  %d/%d  %s  %.0f%%\n\n", completed, total, bar, pct))

	// === SEPARATOR ===
	b.WriteString(sepStyle.Render(strings.Repeat("─", w)) + "\n")

	// === OBJECTIVES ===
	visibleArea := m.getVisibleArea()
	if visibleArea < 1 {
		visibleArea = 5
	}

	for i := m.scroll; i < total && i < m.scroll+visibleArea; i++ {
		obj := mf.Objectives[i]
		isSelected := i == m.cursor
		isDone := m.prog.IsCompleted(m.lab.Name, i)

		var check string
		if isSelected && isDone {
			check = lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor)).Render("✅")
		} else if isSelected {
			check = checkSelected
		} else if isDone {
			check = checkDone
		} else {
			check = checkEmpty
		}

		name := obj.Name
		if isDone {
			name = objDoneStyle.Render(name)
		} else if isSelected {
			name = objSelectedStyle.Render(name)
		} else {
			name = objNameStyle.Render(name)
		}

		arrow := "  "
		if isSelected {
			arrow = arrowStyle.Render("▸ ")
		}

		cat := ""
		if obj.Category != "" {
			cat = categoryStyle.Render("[" + obj.Category + "]")
		}

		b.WriteString(fmt.Sprintf("  %s%s %s%s\n", arrow, check, name, cat))

		if m.showHints[i] {
			if obj.Hint != "" {
				b.WriteString(hintStyle.Render("  💡 "+obj.Hint) + "\n")
			}
			for _, h := range obj.Hints {
				b.WriteString(hintStyle.Render("  💡 "+h) + "\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(sepStyle.Render(strings.Repeat("─", w)) + "\n")
	b.WriteString(footerStyle.Render(" ↑/↓ navigate  ·  space/enter toggle  ·  h hint  ·  x wipe session  ·  q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m model) viewComplete() string {
	w := m.width
	if w <= 0 {
		w = 80
	}

	lines := m.buildCompleteLines(w)

	var b strings.Builder
	totalLines := len(lines)
	padTop := 0
	if m.height > totalLines {
		padTop = (m.height - totalLines) / 2
	}
	for i := 0; i < padTop; i++ {
		b.WriteString("\n")
	}
	for _, line := range lines {
		b.WriteString(line + "\n")
	}

	return b.String()
}

func (m model) buildCompleteLines(w int) []string {
	sepLine := sepStyle.Copy().Align(lipgloss.Center).Width(w)
	titleLine := titleStyle.Copy().Align(lipgloss.Center).Width(w)
	infoLine := infoStyle.Copy().Align(lipgloss.Center).Width(w)
	tagLine := taglineStyle.Copy().Align(lipgloss.Center).Width(w)
	footerLine := footerStyle.Copy().Align(lipgloss.Center).Width(w)

	sepStr := strings.Repeat("─", min(w, 60))

	var lines []string

	lines = append(lines, "")
	lines = append(lines, sepLine.Render(sepStr))
	lines = append(lines, "")
	lines = append(lines, titleLine.Render("🏆  LAB COMPLETE"))
	lines = append(lines, "")

	completed, attempts := m.prog.LabStats(m.lab.Name)
	total := len(m.lab.Manifest.Objectives)

	lines = append(lines, infoLine.Render(fmt.Sprintf("%s — %d/%d objectives completed", m.lab.Manifest.Name, completed, total)))
	lines = append(lines, tagLine.Render(fmt.Sprintf("Total interactions: %d", attempts)))
	lines = append(lines, "")
	lines = append(lines, sepLine.Render(sepStr))
	lines = append(lines, "")
	lines = append(lines, footerLine.Render("press enter or q to exit"))

	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewLab creates a new lab session model
func NewLab(l *lab.Lab, p *progress.Progress, targetURL string) tea.Model {
	return model{
		lab:       l,
		prog:      p,
		phase:     PhaseWelcome,
		targetURL: targetURL,
		showHints: make(map[int]bool),
	}
}

// Result reports how the interactive session ended.
type Result struct {
	// Wiped is true when the user chose to wipe the session: progress was
	// erased and the lab can no longer be resumed.
	Wiped bool
}

// RunLab starts the TUI lab session and records it as the latest session
// so `hacklab resume` can relaunch it later.
func RunLab(l *lab.Lab, p *progress.Progress, targetURL string) (Result, error) {
	p.StartLab(l.Name)
	_ = p.Save()
	_ = session.Save(l.Name)

	prog := tea.NewProgram(
		NewLab(l, p, targetURL),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	final, err := prog.Run()
	if err != nil {
		return Result{}, err
	}

	if fm, ok := final.(model); ok && fm.wiped {
		return Result{Wiped: true}, nil
	}
	return Result{}, nil
}
