package steps

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type Shell struct {
	cfg *config.InstallConfig
	zsh bool
}

func NewShell(cfg *config.InstallConfig) *Shell {
	return &Shell{cfg: cfg, zsh: cfg.Shell == "zsh"}
}

func (s *Shell) Title() string { return "Shell" }

func (s *Shell) Init() tea.Cmd { return nil }

func (s *Shell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "left", "right", "h", "l", "tab":
			s.zsh = !s.zsh
		case "enter":
			if s.zsh {
				s.cfg.Shell = "zsh"
			} else {
				s.cfg.Shell = "bash"
			}
			return s, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	return s, nil
}

func (s *Shell) View() string {
	bash := "  Bash  "
	zsh := "  Zsh  "
	if s.zsh {
		zsh = tui.ActiveStyle.Render("[ Zsh ]")
	} else {
		bash = tui.ActiveStyle.Render("[ Bash ]")
	}
	return "Select default shell for '" + s.cfg.Username + "':\n\n" + bash + "   " + zsh + "\n\n" +
		tui.MutedStyle.Render("Use arrow keys to toggle")
}
