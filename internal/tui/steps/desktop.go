package steps

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type desktopItem struct {
	de   config.DesktopEnvironment
	desc string
}

func (d desktopItem) Title() string       { return d.de.String() }
func (d desktopItem) Description() string { return d.desc }
func (d desktopItem) FilterValue() string { return d.de.String() }

type Desktop struct {
	cfg  *config.InstallConfig
	list list.Model
}

func NewDesktop(cfg *config.InstallConfig) *Desktop {
	items := []list.Item{
		desktopItem{de: config.DesktopGNOME, desc: "Full GNOME desktop with GDM"},
		desktopItem{de: config.DesktopGNOMEMinimal, desc: "GNOME shell, settings, and alacritty"},
		desktopItem{de: config.DesktopKDE, desc: "KDE Plasma desktop with SDDM"},
		desktopItem{de: config.DesktopHyprland, desc: "Hyprland tiling compositor with SDDM"},
		desktopItem{de: config.DesktopNone, desc: "No desktop environment (TTY only)"},
	}
	l := list.New(items, list.NewDefaultDelegate(), 60, 14)
	l.Title = "Select desktop environment"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	if cfg.DesktopSet {
		for i, item := range items {
			if item.(desktopItem).de == cfg.Desktop {
				l.Select(i)
				break
			}
		}
	}
	return &Desktop{cfg: cfg, list: l}
}

func (d *Desktop) Title() string { return "Desktop Environment" }

func (d *Desktop) Init() tea.Cmd { return nil }

func (d *Desktop) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			if item, ok := d.list.SelectedItem().(desktopItem); ok {
				d.cfg.Desktop = item.de
				return d, func() tea.Msg { return tui.SubmitMsg{} }
			}
		}
	}
	var cmd tea.Cmd
	d.list, cmd = d.list.Update(msg)
	return d, cmd
}

func (d *Desktop) View() string {
	return d.list.View()
}
