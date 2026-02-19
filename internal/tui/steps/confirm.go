package steps

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type Confirm struct {
	cfg *config.InstallConfig
}

func NewConfirm(cfg *config.InstallConfig) *Confirm {
	return &Confirm{cfg: cfg}
}

func (c *Confirm) Title() string { return "Confirm Installation" }

func (c *Confirm) Init() tea.Cmd { return nil }

func (c *Confirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			return c, func() tea.Msg { return tui.StartInstallMsg{} }
		}
	}
	return c, nil
}

func (c *Confirm) View() string {
	s := c.cfg.Summary() + "\n"
	s += tui.ErrorStyle.Render("WARNING: This will ERASE ALL DATA on " + c.cfg.Device.Path()) + "\n\n"
	s += tui.MutedStyle.Render("Press Enter to begin installation, Esc to go back.")
	return s
}
