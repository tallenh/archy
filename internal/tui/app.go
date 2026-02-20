package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tallenh/archy/internal/config"
)

// SubmitMsg signals that the current step is done and wants to advance.
type SubmitMsg struct{}

// StartInstallMsg signals that the confirm screen wants to begin installation.
type StartInstallMsg struct{}

// Model is the root Bubble Tea model for the wizard.
type Model struct {
	steps   []StepModel
	current Step
	config  *config.InstallConfig
	keys    KeyMap
	width   int
	height  int
	quitting bool
}

// NewModel creates the root model. The steps slice is populated by the caller
// after all step models are constructed, so we accept it as a parameter.
func NewModel(cfg *config.InstallConfig, steps []StepModel) Model {
	return Model{
		steps:  steps,
		config: cfg,
		keys:   DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	if len(m.steps) > 0 {
		return m.steps[0].Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward to current step
		if int(m.current) < len(m.steps) {
			updated, cmd := m.steps[m.current].Update(msg)
			m.steps[m.current] = updated.(StepModel)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Back):
			if m.current > StepWelcome && m.current != StepInstall {
				return m.prevStep()
			}
		}

	case SubmitMsg:
		return m.nextStep()

	case StartInstallMsg:
		return m.nextStep()
	}

	// Forward to current step
	if int(m.current) < len(m.steps) {
		updated, cmd := m.steps[m.current].Update(msg)
		m.steps[m.current] = updated.(StepModel)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if int(m.current) >= len(m.steps) {
		return ""
	}

	step := m.steps[m.current]
	title := TitleStyle.Render(step.Title())
	body := step.View()
	help := HelpStyle.Render(m.helpText())

	content := lipgloss.JoinVertical(lipgloss.Left, title, body, help)
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (m Model) helpText() string {
	if m.current == StepInstall {
		return "ctrl+c quit"
	}
	if m.current == StepWelcome {
		return "enter next • ctrl+c quit"
	}
	return "enter next • esc back • ctrl+c quit"
}

func (m Model) nextStep() (tea.Model, tea.Cmd) {
	next := m.current + 1
	for int(next) < len(m.steps) {
		if !m.shouldSkip(next) {
			break
		}
		next++
	}
	if int(next) >= len(m.steps) {
		return m, tea.Quit
	}
	m.current = next
	return m, m.steps[m.current].Init()
}

func (m Model) prevStep() (tea.Model, tea.Cmd) {
	prev := m.current - 1
	for prev > StepWelcome {
		if !m.shouldSkip(prev) {
			break
		}
		prev--
	}
	if prev < StepWelcome {
		prev = StepWelcome
	}
	m.current = prev
	return m, m.steps[m.current].Init()
}

// shouldSkip returns true if the given step should be auto-advanced past.
func (m Model) shouldSkip(step Step) bool {
	// Passphrase is always skipped when encryption is disabled
	if step == StepPassphrase && !m.config.Encrypt {
		return true
	}

	// Password steps are skipped in both modes when env var provided the value
	if step == StepUserPassword && m.config.UserPassword != "" {
		return true
	}
	if step == StepRootPassword && m.config.RootPassword != "" {
		return true
	}
	if step == StepPassphrase && m.config.LUKSPassphrase != "" {
		return true
	}

	// Remaining skip logic only applies in skip mode
	if m.config.Mode != "skip" {
		return false
	}

	cfg := m.config
	switch step {
	case StepWelcome, StepConfirm, StepInstall:
		return false
	case StepDevice:
		return cfg.Device.Name != ""
	case StepPartSize:
		return cfg.EFISize != ""
	case StepEncrypt:
		return cfg.EncryptSet
	case StepHostname:
		return cfg.Hostname != ""
	case StepTimezone:
		return cfg.Timezone != ""
	case StepUsername:
		return cfg.Username != ""
	case StepZRAMSize:
		return cfg.ZRAMSize != ""
	case StepDesktop:
		return cfg.DesktopSet
	case StepShell:
		return cfg.Shell != ""
	}
	return false
}
