package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hacklab/internal/lab"
	"hacklab/internal/progress"
)

// Styles
var (
	bgColor       = "#1a1a2e"
	accentColor   = "#e94560"
	cyanColor     = "#00d4ff"
	greenColor    = "#0f0"
	yellowColor   = "#f0e68c"
	dimColor      = "#666"

	bgStyle    = lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor)).Bold(true)
	subStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(dimColor))
	cardStyle  = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(accentColor)).
			Padding(1, 2).
			MarginTop(1)
	objStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor)).Bold(true)
	doneStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor)).Bold(true)
	flagStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(greenColor)).PaddingLeft(2)
	hintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(yellowColor)).Italic(true)
	inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor))
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor)).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor)).Bold(true)
	urlStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(cyanColor)).Underline(true)
)

// Phase represents the current state
type Phase int

const (
	PhaseWelcome Phase = iota
	PhaseMenu
	PhaseHint
	PhaseSubmit
	PhaseResult
	PhaseComplete
)

type model struct {
	lab      *lab.Lab
	progress *progress.Progress
	phase    Phase
	input    string
	cursor   bool
	errMsg   string
	msg      string
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(tickDuration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

const tickDuration = 500 * time.Millisecond

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.phase != PhaseMenu {
				m.phase = PhaseMenu
				m.input = ""
				m.errMsg = ""
				m.msg = ""
				return m, nil
			}
			return m, tea.Quit
		}

		switch m.phase {
		case PhaseWelcome:
			if msg.String() == "enter" || msg.String() == " " {
				m.phase = PhaseMenu
				m.errMsg = ""
				m.msg = ""
			}

		case PhaseMenu:
			switch msg.String() {
			case "enter":
				if m.input == "" {
					m.errMsg = "type a command: objectives, hint N, submit FLAG, url, quit"
					return m, nil
				}
				return m.processCommand(m.input)
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
			m.errMsg = ""

		case PhaseHint:
			if msg.String() == "enter" {
				m.phase = PhaseMenu
				m.input = ""
			}

		case PhaseSubmit:
			if msg.String() == "enter" {
				m.phase = PhaseMenu
				m.input = ""
			}

		case PhaseResult:
			if msg.String() == "enter" || msg.String() == " " {
				m.phase = PhaseMenu
				m.input = ""
			}

		case PhaseComplete:
			if msg.String() == "q" || msg.String() == "enter" {
				return m, tea.Quit
			}
		}

	case tickMsg:
		m.cursor = !m.cursor
		return m, tickCmd()
	}
	return m, nil
}

func (m *model) processCommand(input string) (tea.Model, tea.Cmd) {
	input = strings.TrimSpace(input)
	parts := strings.SplitN(input, " ", 2)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "objectives", "obj", "o", "":
		m.phase = PhaseMenu
		m.msg = ""
		return m, nil

	case "hint", "h":
		if len(parts) < 2 {
			m.errMsg = "usage: hint <number>"
			return m, nil
		}
		var idx int
		fmt.Sscanf(parts[1], "%d", &idx)
		idx-- // 1-indexed
		if idx < 0 || idx >= len(m.lab.Manifest.Objectives) {
			m.errMsg = fmt.Sprintf("objective %d doesn't exist (1-%d)", idx+1, len(m.lab.Manifest.Objectives))
			return m, nil
		}
		obj := m.lab.Manifest.Objectives[idx]
		m.phase = PhaseHint
		m.msg = formatHint(obj, idx+1)
		return m, nil

	case "submit", "s", "flag", "f":
		if len(parts) < 2 {
			m.errMsg = "usage: submit <flag>"
			return m, nil
		}
		flag := strings.TrimSpace(parts[1])
		return m.checkFlag(flag)

	case "url", "target", "t":
		m.phase = PhaseMenu
		if m.lab.Manifest.Image != "" {
			m.msg = fmt.Sprintf("Target URL: http://localhost:%d", m.lab.Manifest.Port)
		} else {
			m.msg = "This lab uses docker-compose — check the output above for endpoints"
		}
		return m, nil

	case "quit", "exit", "q":
		return m, tea.Quit

	default:
		m.errMsg = fmt.Sprintf("unknown command '%s' — try: objectives, hint N, submit FLAG, url, quit", cmd)
		return m, nil
	}
}

func (m *model) checkFlag(flag string) (tea.Model, tea.Cmd) {
	// Check against all uncompleted objectives
	for i, obj := range m.lab.Manifest.Objectives {
		if m.progress.IsCompleted(m.lab.Name, i) {
			continue
		}
		if strings.EqualFold(flag, obj.Flag) {
			m.progress.RecordAttempt(m.lab.Name, i)
			m.progress.CompleteObjective(m.lab.Name, i)
			_ = m.progress.Save()

			// Check if all done
			if len(m.progress.Labs[m.lab.Name].Completed) == len(m.lab.Manifest.Objectives) {
				m.phase = PhaseComplete
				m.msg = "all objectives complete"
				return m, nil
			}

			m.phase = PhaseResult
			m.msg = fmt.Sprintf("✅ Correct! '%s' completed", obj.Name)
			return m, nil
		}
	}

	// Wrong flag
	m.progress.RecordAttempt(m.lab.Name, -1) // track failed attempts
	_ = m.progress.Save()
	m.phase = PhaseResult
	m.msg = "❌ Incorrect flag. Try again or request a hint."
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("\n")

	switch m.phase {
	case PhaseWelcome:
		b.WriteString(m.viewWelcome())
	case PhaseMenu:
		b.WriteString(m.viewMenu())
	case PhaseHint:
		b.WriteString(m.viewMenu())
		b.WriteString(m.msg)
		b.WriteString("\n\n")
	case PhaseResult:
		b.WriteString(m.viewMenu())
		b.WriteString(m.msg)
		b.WriteString("\n\n")
	case PhaseComplete:
		b.WriteString(m.viewComplete())
	}

	b.WriteString("\n")
	return bgStyle.Render(b.String())
}

func (m model) viewWelcome() string {
	mf := m.lab.Manifest
	var b strings.Builder

	b.WriteString(titleStyle.Render("⚡ HACKLAB") + "\n")
	b.WriteString(subStyle.Render("your terminal hacking playground") + "\n\n")

	b.WriteString(cardStyle.Render(
		fmt.Sprintf("🎯 Lab: %s\n", mf.Name) +
			fmt.Sprintf("📊 Difficulty: %s\n", capitalize(mf.Difficulty)) +
			fmt.Sprintf("📝 Objectives: %d\n", len(mf.Objectives)) +
			func() string {
				if mf.Description != "" {
					return "\n" + subStyle.Render(mf.Description)
				}
				return ""
			}(),
	) + "\n\n")

	completed, _ := m.progress.LabStats(m.lab.Name)
	if completed > 0 {
		b.WriteString(subStyle.Render(fmt.Sprintf("  (previously completed: %d/%d)", completed, len(mf.Objectives))) + "\n\n")
	}

	b.WriteString(subStyle.Render("  type commands to interact • q to quit") + "\n")
	b.WriteString(subStyle.Render("  press enter to begin") + "\n")

	return b.String()
}

func (m model) viewMenu() string {
	mf := m.lab.Manifest
	var b strings.Builder

	// Header
	if mf.Image != "" {
		b.WriteString(infoStyle.Render(fmt.Sprintf("  📡 Target: http://localhost:%d", mf.Port)) + "\n")
	}
	b.WriteString("\n")

	// Objectives list
	b.WriteString(titleStyle.Render("  OBJECTIVES") + "\n\n")

	for i, obj := range m.lab.Manifest.Objectives {
		// Check if completed
		done := m.progress.IsCompleted(m.lab.Name, i)
		idx := i + 1

		status := "  "
		if done {
			status = doneStyle.Render("  ✅")
		} else {
			status = subStyle.Render("  ◻ ")
		}

		name := objStyle.Render(fmt.Sprintf("%d. %s", idx, obj.Name))
		if done {
			name = doneStyle.Render(fmt.Sprintf("%d. %s", idx, obj.Name))
		}

		cat := ""
		if obj.Category != "" {
			cat = " " + subStyle.Render("["+obj.Category+"]")
		}

		b.WriteString(fmt.Sprintf("%s %s%s\n", status, name, cat))
	}

	b.WriteString("\n")
	b.WriteString(subStyle.Render("  ─────────────────────────────────") + "\n")
	b.WriteString("\n")

	// Error message
	if m.errMsg != "" {
		b.WriteString(errStyle.Render("  "+m.errMsg) + "\n\n")
	}

	// Command input
	b.WriteString(promptStyle.Render("  ❯ ") + inputStyle.Render(m.input) + cursor(m.cursor))
	b.WriteString("\n")

	// Help
	b.WriteString("\n")
	b.WriteString(subStyle.Render("  commands: objectives | hint N | submit FLAG | url | quit") + "\n")

	return b.String()
}

func (m model) viewComplete() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🏆 LAB COMPLETE!") + "\n\n")

	completed, attempts := m.progress.LabStats(m.lab.Name)
	total := len(m.lab.Manifest.Objectives)

	b.WriteString(cardStyle.Render(
		fmt.Sprintf("Lab: %s\n", m.lab.Manifest.Name) +
			fmt.Sprintf("Completed: %d/%d objectives\n", completed, total) +
			fmt.Sprintf("Total attempts: %d\n", attempts),
	) + "\n\n")

	b.WriteString(subStyle.Render("  press enter or q to exit") + "\n")

	return b.String()
}

func formatHint(obj lab.Objective, idx int) string {
	var b strings.Builder
	b.WriteString(hintStyle.Render(fmt.Sprintf("  💡 Hint for '%s':", obj.Name)) + "\n\n")

	if obj.Hint != "" {
		b.WriteString(flagStyle.Render(obj.Hint) + "\n")
	}
	for _, h := range obj.Hints {
		b.WriteString(flagStyle.Render(h) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(subStyle.Render("  press enter to continue") + "\n")
	return b.String()
}

func cursor(visible bool) string {
	if visible {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor)).Render("█")
	}
	return " "
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// NewLab creates a new lab session model
func NewLab(l *lab.Lab, p *progress.Progress, targetURL string) tea.Model {
	return model{
		lab:      l,
		progress: p,
		phase:    PhaseWelcome,
	}
}

// RunLab starts the TUI lab session
func RunLab(l *lab.Lab, p *progress.Progress, targetURL string) error {
	p.StartLab(l.Name)
	_ = p.Save()

	prog := tea.NewProgram(NewLab(l, p, targetURL), tea.WithAltScreen())
	_, err := prog.Run()
	return err
}
