package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hacklab/internal/lab"
	"hacklab/internal/progress"
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
// HACKLAB ASCII logo — pure ASCII, no Unicode box-drawing that causes terminal width drift
const asciiLogo = `H   H  AAAAA  CCCC   K   K  L      AAAAA  BBBB
H   H  A   A  C      K  K   L      A   A  B   B
HHHHH  AAAAA  C      KKK    L      AAAAA  BBBB
H   H  A   A  C      K  K   L      A   A  B   B
H   H  A   A  CCCC   K   K  LLLLL  A   A  BBBB`

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
)

type model struct {
	lab       *lab.Lab
	prog      *progress.Progress
	phase     Phase
	cursor    int
	scroll    int
	width     int
	height    int
	targetURL string
	showHints map[int]bool
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
			if msg.String() == "enter" || msg.String() == " " {
				m.phase = PhaseQuiz
				m.cursor = 0
				m.scroll = 0
				m.showHints = make(map[int]bool)
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
			case "q":
				return m, tea.Quit
			}

		case PhaseComplete:
			if msg.String() == "enter" || msg.String() == " " || msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
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
	default:
		return ""
	}
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
	lines = append(lines, footerLine.Render("press enter to begin  ·  q to quit"))

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
	b.WriteString(footerStyle.Render(" ↑/↓ navigate  ·  space/enter toggle  ·  h hint  ·  q quit"))
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

// RunLab starts the TUI lab session
func RunLab(l *lab.Lab, p *progress.Progress, targetURL string) error {
	p.StartLab(l.Name)
	_ = p.Save()

	prog := tea.NewProgram(
		NewLab(l, p, targetURL),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := prog.Run()
	return err
}
