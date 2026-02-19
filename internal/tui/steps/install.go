package steps

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/installer"
	"github.com/tallenh/archy/internal/tui"
)

// PhaseUpdateMsg carries an installer progress update into the TUI.
type PhaseUpdateMsg installer.PhaseUpdate

type Install struct {
	cfg      *config.InstallConfig
	spinner  spinner.Model
	progress progress.Model
	logs     []string
	phase    string
	percent  float64
	done     bool
	err      error
	sub      <-chan installer.PhaseUpdate
}

func NewInstall(cfg *config.InstallConfig) *Install {
	s := spinner.New()
	s.Spinner = spinner.Dot

	p := progress.New(progress.WithDefaultGradient())
	p.Width = 50

	return &Install{
		cfg:      cfg,
		spinner:  s,
		progress: p,
	}
}

func (i *Install) Title() string { return "Installing Arch Linux" }

func (i *Install) Init() tea.Cmd {
	ch := make(chan installer.PhaseUpdate, 20)
	i.sub = ch

	inst := installer.New(i.cfg, ch)
	go func() {
		inst.Run()
		close(ch)
	}()

	return tea.Batch(i.spinner.Tick, i.waitForUpdate())
}

func (i *Install) waitForUpdate() tea.Cmd {
	return func() tea.Msg {
		if i.sub == nil {
			return nil
		}
		update, ok := <-i.sub
		if !ok {
			return nil
		}
		return PhaseUpdateMsg(update)
	}
}

func (i *Install) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PhaseUpdateMsg:
		i.phase = msg.Description
		i.percent = msg.Percent
		if msg.LogLine != "" {
			i.logs = append(i.logs, msg.LogLine)
		}
		if msg.Err != nil {
			i.err = msg.Err
			i.done = true
			return i, nil
		}
		if msg.Done {
			i.done = true
			return i, nil
		}
		return i, tea.Batch(i.spinner.Tick, i.waitForUpdate())
	case spinner.TickMsg:
		var cmd tea.Cmd
		i.spinner, cmd = i.spinner.Update(msg)
		return i, cmd
	}
	return i, nil
}

func (i *Install) View() string {
	var b strings.Builder

	if !i.done {
		fmt.Fprintf(&b, "%s %s\n\n", i.spinner.View(), i.phase)
		b.WriteString(i.progress.ViewAs(i.percent) + "\n\n")
	} else if i.err != nil {
		fmt.Fprintf(&b, "%s\n\n", tui.ErrorStyle.Render("Installation failed: "+i.err.Error()))
		b.WriteString(i.progress.ViewAs(i.percent) + "\n\n")
	} else {
		b.WriteString(tui.SuccessStyle.Render("Installation complete!") + "\n\n")
		b.WriteString(i.progress.ViewAs(1.0) + "\n\n")
		b.WriteString("Remove the installation media and reboot.\n")
	}

	// Show last 10 log lines
	start := 0
	if len(i.logs) > 10 {
		start = len(i.logs) - 10
	}
	for _, line := range i.logs[start:] {
		b.WriteString(tui.MutedStyle.Render(line) + "\n")
	}

	return b.String()
}
