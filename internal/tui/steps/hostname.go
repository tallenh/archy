package steps

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type Hostname struct {
	cfg   *config.InstallConfig
	input textinput.Model
	err   string
}

func NewHostname(cfg *config.InstallConfig) *Hostname {
	ti := textinput.New()
	ti.Placeholder = "archlinux"
	ti.CharLimit = 63
	ti.Width = 40
	if cfg.Hostname != "" {
		ti.SetValue(cfg.Hostname)
	}
	return &Hostname{cfg: cfg, input: ti}
}

func (h *Hostname) Title() string { return "Hostname" }

func (h *Hostname) Init() tea.Cmd { return h.input.Focus() }

func (h *Hostname) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			val := h.input.Value()
			if err := config.ValidateHostname(val); err != nil {
				h.err = err.Error()
				return h, nil
			}
			h.err = ""
			h.cfg.Hostname = val
			return h, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	var cmd tea.Cmd
	h.input, cmd = h.input.Update(msg)
	return h, cmd
}

func (h *Hostname) View() string {
	s := "Enter a hostname for this machine:\n\n" + h.input.View()
	if h.err != "" {
		s += "\n" + tui.ErrorStyle.Render(h.err)
	}
	return s
}
