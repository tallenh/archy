package steps

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/tui"
)

type Docker struct {
	cfg         *config.InstallConfig
	docker      bool
	dockerGroup bool
	focus       int // 0 = docker toggle, 1 = docker group toggle
}

func NewDocker(cfg *config.InstallConfig) *Docker {
	return &Docker{
		cfg:         cfg,
		docker:      cfg.Docker,
		dockerGroup: cfg.DockerGroup,
	}
}

func (d *Docker) Title() string { return "Docker" }

func (d *Docker) Init() tea.Cmd { return nil }

func (d *Docker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "left", "right", "h", "l", "tab":
			if d.focus == 0 {
				d.docker = !d.docker
				if !d.docker {
					d.focus = 0
				}
			} else {
				d.dockerGroup = !d.dockerGroup
			}
		case "up", "k":
			if d.focus > 0 {
				d.focus--
			}
		case "down", "j":
			if d.docker && d.focus < 1 {
				d.focus++
			}
		case "y":
			if d.focus == 0 {
				d.docker = true
			} else {
				d.dockerGroup = true
			}
		case "n":
			if d.focus == 0 {
				d.docker = false
				d.focus = 0
			} else {
				d.dockerGroup = false
			}
		case "enter":
			d.cfg.Docker = d.docker
			d.cfg.DockerGroup = d.dockerGroup
			return d, func() tea.Msg { return tui.SubmitMsg{} }
		}
	}
	return d, nil
}

func (d *Docker) View() string {
	// Row 1: Install Docker?
	yes := "  Yes  "
	no := "  No  "
	if d.docker {
		if d.focus == 0 {
			yes = tui.ActiveStyle.Render("[ Yes ]")
		} else {
			yes = tui.SelectedStyle.Render("[ Yes ]")
		}
	} else {
		if d.focus == 0 {
			no = tui.ActiveStyle.Render("[ No ]")
		} else {
			no = tui.SelectedStyle.Render("[ No ]")
		}
	}
	out := "Install Docker?\n\n" + yes + "   " + no + "\n"

	// Row 2: Add user to docker group? (only when docker is enabled)
	if d.docker {
		gYes := "  Yes  "
		gNo := "  No  "
		if d.dockerGroup {
			if d.focus == 1 {
				gYes = tui.ActiveStyle.Render("[ Yes ]")
			} else {
				gYes = tui.SelectedStyle.Render("[ Yes ]")
			}
		} else {
			if d.focus == 1 {
				gNo = tui.ActiveStyle.Render("[ No ]")
			} else {
				gNo = tui.SelectedStyle.Render("[ No ]")
			}
		}
		out += "\nAdd user to docker group?\n\n" + gYes + "   " + gNo + "\n"
	}

	out += "\n" + tui.MutedStyle.Render("Use arrow keys or y/n to toggle, up/down to switch rows")
	return out
}
