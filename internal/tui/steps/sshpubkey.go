package steps

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type SSHPubKey struct {
	cfg       *config.InstallConfig
	input     textinput.Model
	err       string
	fromConfig bool // true when key was loaded from config (APPROVE mode)
}

func NewSSHPubKey(cfg *config.InstallConfig) *SSHPubKey {
	ti := textinput.New()
	fromConfig := cfg.SSHPubKeyFromConfig

	if fromConfig {
		ti.Placeholder = "Type APPROVE to install, or press Enter to skip"
		ti.CharLimit = 7
		ti.Width = 40
	} else {
		ti.Placeholder = "ssh-ed25519 AAAA... user@host"
		ti.CharLimit = 1024
		ti.Width = 72
	}

	return &SSHPubKey{cfg: cfg, input: ti, fromConfig: fromConfig}
}

func (s *SSHPubKey) Title() string { return "SSH Public Key" }

func (s *SSHPubKey) Init() tea.Cmd { return s.input.Focus() }

func (s *SSHPubKey) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			val := strings.TrimSpace(s.input.Value())

			if s.fromConfig {
				return s.handleConfigMode(val)
			}
			return s.handleInteractiveMode(val)
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s *SSHPubKey) handleConfigMode(val string) (tea.Model, tea.Cmd) {
	if val == "" {
		// Skip — clear the config key
		s.err = ""
		s.cfg.SSHPubKey = ""
		return s, func() tea.Msg { return tui.SubmitMsg{} }
	}
	if val == "APPROVE" {
		// Keep the config key as-is
		s.err = ""
		return s, func() tea.Msg { return tui.SubmitMsg{} }
	}
	s.err = "Type APPROVE to install this key, or press Enter with empty input to skip"
	return s, nil
}

func (s *SSHPubKey) handleInteractiveMode(val string) (tea.Model, tea.Cmd) {
	if val == "" {
		// Empty is allowed — skip pubkey
		s.err = ""
		s.cfg.SSHPubKey = ""
		return s, func() tea.Msg { return tui.SubmitMsg{} }
	}
	if err := config.ValidateSSHPubKey(val); err != nil {
		s.err = err.Error()
		return s, nil
	}
	s.err = ""
	s.cfg.SSHPubKey = val
	return s, func() tea.Msg { return tui.SubmitMsg{} }
}

func (s *SSHPubKey) View() string {
	var b strings.Builder

	if s.fromConfig {
		b.WriteString("The configuration file specifies an SSH public key to install\n")
		b.WriteString("for '" + s.cfg.Username + "':\n\n")
		b.WriteString(tui.MutedStyle.Render(s.cfg.SSHPubKey) + "\n\n")
		b.WriteString(s.input.View())
		if s.err != "" {
			b.WriteString("\n" + tui.ErrorStyle.Render(s.err))
		}
		b.WriteString("\n" + tui.MutedStyle.Render("Type APPROVE to install this key, or press Enter to skip."))
	} else {
		b.WriteString("Paste an SSH public key for '" + s.cfg.Username + "':\n\n")
		b.WriteString(s.input.View())
		if s.err != "" {
			b.WriteString("\n" + tui.ErrorStyle.Render(s.err))
		}
		b.WriteString("\n" + tui.MutedStyle.Render("Added to ~/.ssh/authorized_keys. Press Enter to skip."))
	}

	return b.String()
}
