package steps

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type tzItem string

func (t tzItem) Title() string       { return string(t) }
func (t tzItem) Description() string { return "" }
func (t tzItem) FilterValue() string { return string(t) }

type Timezone struct {
	cfg  *config.InstallConfig
	list list.Model
}

func NewTimezone(cfg *config.InstallConfig, zones []string) *Timezone {
	items := make([]list.Item, len(zones))
	for i, z := range zones {
		items[i] = tzItem(z)
	}
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New(items, delegate, 60, 20)
	l.Title = "Select timezone"
	l.SetShowHelp(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	return &Timezone{cfg: cfg, list: l}
}

func (t *Timezone) Title() string { return "Timezone" }

func (t *Timezone) Init() tea.Cmd { return nil }

func (t *Timezone) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" && !t.list.SettingFilter() {
			if item, ok := t.list.SelectedItem().(tzItem); ok {
				t.cfg.Timezone = string(item)
				return t, func() tea.Msg { return tui.SubmitMsg{} }
			}
		}
	}
	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return t, cmd
}

func (t *Timezone) View() string {
	return t.list.View()
}
