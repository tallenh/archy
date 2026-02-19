package steps

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/system"
	"github.com/tallenh/archy/internal/tui"
)

type PartSize struct {
	cfg   *config.InstallConfig
	input textinput.Model
	err   string
}

func NewPartSize(cfg *config.InstallConfig) *PartSize {
	ti := textinput.New()
	ti.Placeholder = "512M"
	ti.SetValue(cfg.EFISize)
	ti.CharLimit = 10
	ti.Width = 20
	return &PartSize{cfg: cfg, input: ti}
}

func (p *PartSize) Title() string { return "EFI Partition Size" }

func (p *PartSize) Init() tea.Cmd { return p.input.Focus() }

func (p *PartSize) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			val := p.input.Value()
			if val == "" {
				val = "512M"
			}
			if err := system.ValidatePartitionSize(val); err != nil {
				p.err = err.Error()
				return p, nil
			}
			p.err = ""
			p.cfg.EFISize = val
			return p, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return p, cmd
}

func (p *PartSize) View() string {
	s := "Enter the size of the EFI partition:\n\n" + p.input.View()
	if p.err != "" {
		s += "\n" + tui.ErrorStyle.Render(p.err)
	}
	return s
}
