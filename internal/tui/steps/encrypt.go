package steps

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type Encrypt struct {
	cfg     *config.InstallConfig
	encrypt bool
}

func NewEncrypt(cfg *config.InstallConfig) *Encrypt {
	return &Encrypt{cfg: cfg}
}

func (e *Encrypt) Title() string { return "Disk Encryption" }

func (e *Encrypt) Init() tea.Cmd { return nil }

func (e *Encrypt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "left", "right", "h", "l", "tab":
			e.encrypt = !e.encrypt
		case "y":
			e.encrypt = true
		case "n":
			e.encrypt = false
		case "enter":
			e.cfg.Encrypt = e.encrypt
			return e, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	return e, nil
}

func (e *Encrypt) View() string {
	yes := "  Yes  "
	no := "  No  "
	if e.encrypt {
		yes = tui.ActiveStyle.Render("[ Yes ]")
		no = "  No  "
	} else {
		yes = "  Yes  "
		no = tui.ActiveStyle.Render("[ No ]")
	}
	return "Enable LUKS2 disk encryption?\n\n" + yes + "   " + no + "\n\n" +
		tui.MutedStyle.Render("Use arrow keys or y/n to toggle")
}
