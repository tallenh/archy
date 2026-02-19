package steps

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/system"
	"github.com/tallenh/archy/internal/tui"
)

type Passphrase struct {
	cfg     *config.InstallConfig
	pass    textinput.Model
	confirm textinput.Model
	focused int // 0 = pass, 1 = confirm
	err     string
}

func NewPassphrase(cfg *config.InstallConfig) *Passphrase {
	pass := textinput.New()
	pass.Placeholder = "Passphrase"
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '*'
	pass.CharLimit = 128
	pass.Width = 40

	confirm := textinput.New()
	confirm.Placeholder = "Confirm passphrase"
	confirm.EchoMode = textinput.EchoPassword
	confirm.EchoCharacter = '*'
	confirm.CharLimit = 128
	confirm.Width = 40

	return &Passphrase{cfg: cfg, pass: pass, confirm: confirm}
}

func (p *Passphrase) Title() string { return "LUKS Passphrase" }

func (p *Passphrase) Init() tea.Cmd { return p.pass.Focus() }

func (p *Passphrase) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "tab", "shift+tab":
			if p.focused == 0 {
				p.focused = 1
				p.pass.Blur()
				return p, p.confirm.Focus()
			}
			p.focused = 0
			p.confirm.Blur()
			return p, p.pass.Focus()
		case "enter":
			if p.focused == 0 {
				p.focused = 1
				p.pass.Blur()
				return p, p.confirm.Focus()
			}
			if err := system.ValidatePassphrase(p.pass.Value()); err != nil {
				p.err = err.Error()
				return p, nil
			}
			if p.pass.Value() != p.confirm.Value() {
				p.err = "passphrases do not match"
				return p, nil
			}
			p.err = ""
			p.cfg.LUKSPassphrase = p.pass.Value()
			return p, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}

	var cmd tea.Cmd
	if p.focused == 0 {
		p.pass, cmd = p.pass.Update(msg)
	} else {
		p.confirm, cmd = p.confirm.Update(msg)
	}
	return p, cmd
}

func (p *Passphrase) View() string {
	s := "Enter encryption passphrase:\n\n"
	s += p.pass.View() + "\n"
	s += p.confirm.View()
	if p.err != "" {
		s += "\n" + tui.ErrorStyle.Render(p.err)
	}
	s += "\n" + tui.MutedStyle.Render("Tab to switch fields")
	return s
}
