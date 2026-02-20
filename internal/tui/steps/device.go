package steps

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type deviceItem struct {
	device config.BlockDevice
}

func (d deviceItem) Title() string       { return d.device.Path() }
func (d deviceItem) Description() string { return d.device.Size + "  " + d.device.Model }
func (d deviceItem) FilterValue() string { return d.device.Name }

type Device struct {
	cfg  *config.InstallConfig
	list list.Model
}

func NewDevice(cfg *config.InstallConfig, disks []config.BlockDevice) *Device {
	items := make([]list.Item, len(disks))
	selectedIdx := 0
	for i, d := range disks {
		items[i] = deviceItem{device: d}
		if cfg.Device.Name != "" && d.Name == cfg.Device.Name {
			selectedIdx = i
		}
	}
	l := list.New(items, list.NewDefaultDelegate(), 60, 14)
	l.Title = "Select target disk"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	if cfg.Device.Name != "" {
		l.Select(selectedIdx)
	}
	return &Device{cfg: cfg, list: l}
}

func (d *Device) Title() string { return "Target Device" }

func (d *Device) Init() tea.Cmd { return nil }

func (d *Device) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			if item, ok := d.list.SelectedItem().(deviceItem); ok {
				d.cfg.Device = item.device
				return d, func() tea.Msg { return tui.SubmitMsg{} }
			}
		}
	}
	var cmd tea.Cmd
	d.list, cmd = d.list.Update(msg)
	return d, cmd
}

func (d *Device) View() string {
	return d.list.View()
}
