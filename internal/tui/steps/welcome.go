package steps

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/tui"
)

type Welcome struct{}

func NewWelcome() *Welcome { return &Welcome{} }

func (w *Welcome) Title() string { return "Welcome to Archy" }

func (w *Welcome) Init() tea.Cmd { return nil }

func (w *Welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			return w, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	return w, nil
}

func (w *Welcome) View() string {
	return "Arch Linux Installer\n\n" +
		"This wizard will guide you through installing Arch Linux.\n" +
		"UEFI boot mode with btrfs subvolumes, optional LUKS encryption.\n\n" +
		tui.MutedStyle.Render("Press Enter to begin.")
}
