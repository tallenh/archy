package steps

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/system"
	"github.com/tallenh/archy/internal/tui"
)

type ZRAMSize struct {
	cfg   *config.InstallConfig
	input textinput.Model
	err   string
}

func NewZRAMSize(cfg *config.InstallConfig) *ZRAMSize {
	ti := textinput.New()
	ti.Placeholder = cfg.ZRAMSize
	ti.SetValue(cfg.ZRAMSize)
	ti.CharLimit = 10
	ti.Width = 20
	return &ZRAMSize{cfg: cfg, input: ti}
}

func (z *ZRAMSize) Title() string { return "ZRAM Swap Size" }

func (z *ZRAMSize) Init() tea.Cmd { return z.input.Focus() }

func (z *ZRAMSize) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			val := z.input.Value()
			if val == "" {
				val = z.cfg.ZRAMSize
			}
			if err := system.ValidateZRAMSize(val); err != nil {
				z.err = err.Error()
				return z, nil
			}
			z.err = ""
			z.cfg.ZRAMSize = val
			return z, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	var cmd tea.Cmd
	z.input, cmd = z.input.Update(msg)
	return z, cmd
}

func (z *ZRAMSize) View() string {
	s := "Enter ZRAM swap size (default: half of RAM):\n\n" + z.input.View()
	if z.err != "" {
		s += "\n" + tui.ErrorStyle.Render(z.err)
	}
	return s
}
