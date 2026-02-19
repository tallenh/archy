package steps

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/system"
	"github.com/tallenh/archy/internal/tui"
)

type Username struct {
	cfg   *config.InstallConfig
	input textinput.Model
	err   string
}

func NewUsername(cfg *config.InstallConfig) *Username {
	ti := textinput.New()
	ti.Placeholder = "user"
	ti.CharLimit = 32
	ti.Width = 40
	return &Username{cfg: cfg, input: ti}
}

func (u *Username) Title() string { return "Username" }

func (u *Username) Init() tea.Cmd { return u.input.Focus() }

func (u *Username) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			val := u.input.Value()
			if err := system.ValidateUsername(val); err != nil {
				u.err = err.Error()
				return u, nil
			}
			u.err = ""
			u.cfg.Username = val
			return u, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	var cmd tea.Cmd
	u.input, cmd = u.input.Update(msg)
	return u, cmd
}

func (u *Username) View() string {
	s := "Enter the primary user's username:\n\n" + u.input.View()
	if u.err != "" {
		s += "\n" + tui.ErrorStyle.Render(u.err)
	}
	return s
}
