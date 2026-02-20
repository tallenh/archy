package steps

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type SSHD struct {
	cfg  *config.InstallConfig
	sshd bool
}

func NewSSHD(cfg *config.InstallConfig) *SSHD {
	return &SSHD{cfg: cfg, sshd: cfg.SSHD}
}

func (s *SSHD) Title() string { return "SSH Server" }

func (s *SSHD) Init() tea.Cmd { return nil }

func (s *SSHD) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "left", "right", "h", "l", "tab":
			s.sshd = !s.sshd
		case "y":
			s.sshd = true
		case "n":
			s.sshd = false
		case "enter":
			s.cfg.SSHD = s.sshd
			return s, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	return s, nil
}

func (s *SSHD) View() string {
	yes := "  Yes  "
	no := "  No  "
	if s.sshd {
		yes = tui.ActiveStyle.Render("[ Yes ]")
	} else {
		no = tui.ActiveStyle.Render("[ No ]")
	}
	return "Install and enable SSH server (sshd)?\n\n" + yes + "   " + no + "\n\n" +
		tui.MutedStyle.Render("Root login disabled, password auth disabled, pubkey auth only.\n") +
		tui.MutedStyle.Render("Use arrow keys or y/n to toggle")
}
