package steps

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/system"
	"github.com/tallenh/archy/internal/tui"
)

type UserPassword struct {
	cfg     *config.InstallConfig
	pass    textinput.Model
	confirm textinput.Model
	focused int
	err     string
}

func NewUserPassword(cfg *config.InstallConfig) *UserPassword {
	pass := textinput.New()
	pass.Placeholder = "Password"
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '*'
	pass.CharLimit = 128
	pass.Width = 40

	confirm := textinput.New()
	confirm.Placeholder = "Confirm password"
	confirm.EchoMode = textinput.EchoPassword
	confirm.EchoCharacter = '*'
	confirm.CharLimit = 128
	confirm.Width = 40

	return &UserPassword{cfg: cfg, pass: pass, confirm: confirm}
}

func (u *UserPassword) Title() string { return "User Password" }

func (u *UserPassword) Init() tea.Cmd { return u.pass.Focus() }

func (u *UserPassword) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "tab", "shift+tab":
			if u.focused == 0 {
				u.focused = 1
				u.pass.Blur()
				return u, u.confirm.Focus()
			}
			u.focused = 0
			u.confirm.Blur()
			return u, u.pass.Focus()
		case "enter":
			if u.focused == 0 {
				u.focused = 1
				u.pass.Blur()
				return u, u.confirm.Focus()
			}
			if err := system.ValidatePassword(u.pass.Value()); err != nil {
				u.err = err.Error()
				return u, nil
			}
			if u.pass.Value() != u.confirm.Value() {
				u.err = "passwords do not match"
				return u, nil
			}
			u.err = ""
			u.cfg.UserPassword = u.pass.Value()
			return u, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}

	var cmd tea.Cmd
	if u.focused == 0 {
		u.pass, cmd = u.pass.Update(msg)
	} else {
		u.confirm, cmd = u.confirm.Update(msg)
	}
	return u, cmd
}

func (u *UserPassword) View() string {
	s := "Enter password for user '" + u.cfg.Username + "':\n\n"
	s += u.pass.View() + "\n"
	s += u.confirm.View()
	if u.err != "" {
		s += "\n" + tui.ErrorStyle.Render(u.err)
	}
	s += "\n" + tui.MutedStyle.Render("Tab to switch fields")
	return s
}
