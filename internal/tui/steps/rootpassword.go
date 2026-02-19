package steps

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/system"
	"github.com/tallenh/archy/internal/tui"
)

type RootPassword struct {
	cfg     *config.InstallConfig
	pass    textinput.Model
	confirm textinput.Model
	focused int
	err     string
}

func NewRootPassword(cfg *config.InstallConfig) *RootPassword {
	pass := textinput.New()
	pass.Placeholder = "Root password"
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '*'
	pass.CharLimit = 128
	pass.Width = 40

	confirm := textinput.New()
	confirm.Placeholder = "Confirm root password"
	confirm.EchoMode = textinput.EchoPassword
	confirm.EchoCharacter = '*'
	confirm.CharLimit = 128
	confirm.Width = 40

	return &RootPassword{cfg: cfg, pass: pass, confirm: confirm}
}

func (r *RootPassword) Title() string { return "Root Password" }

func (r *RootPassword) Init() tea.Cmd { return r.pass.Focus() }

func (r *RootPassword) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "tab", "shift+tab":
			if r.focused == 0 {
				r.focused = 1
				r.pass.Blur()
				return r, r.confirm.Focus()
			}
			r.focused = 0
			r.confirm.Blur()
			return r, r.pass.Focus()
		case "enter":
			if r.focused == 0 {
				r.focused = 1
				r.pass.Blur()
				return r, r.confirm.Focus()
			}
			if err := system.ValidatePassword(r.pass.Value()); err != nil {
				r.err = err.Error()
				return r, nil
			}
			if r.pass.Value() != r.confirm.Value() {
				r.err = "passwords do not match"
				return r, nil
			}
			r.err = ""
			r.cfg.RootPassword = r.pass.Value()
			return r, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}

	var cmd tea.Cmd
	if r.focused == 0 {
		r.pass, cmd = r.pass.Update(msg)
	} else {
		r.confirm, cmd = r.confirm.Update(msg)
	}
	return r, cmd
}

func (r *RootPassword) View() string {
	s := "Enter root password:\n\n"
	s += r.pass.View() + "\n"
	s += r.confirm.View()
	if r.err != "" {
		s += "\n" + tui.ErrorStyle.Render(r.err)
	}
	s += "\n" + tui.MutedStyle.Render("Tab to switch fields")
	return s
}
