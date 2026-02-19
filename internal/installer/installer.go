package installer

import (
	"fmt"
	"os/exec"

	"github.com/tallenh/archy/internal/config"
)

// Installer orchestrates the installation process.
type Installer struct {
	cfg      *config.InstallConfig
	progress chan<- PhaseUpdate
}

// New creates an Installer that reports progress to the given channel.
func New(cfg *config.InstallConfig, progress chan<- PhaseUpdate) *Installer {
	return &Installer{cfg: cfg, progress: progress}
}

// Run executes all installation phases in order.
func (inst *Installer) Run() {
	phases := []struct {
		phase Phase
		fn    func() error
		skip  bool
	}{
		{PhasePrepare, inst.prepare, false},
		{PhasePartition, inst.partition, false},
		{PhaseLUKS, inst.setupLUKS, !inst.cfg.Encrypt},
		{PhaseBtrfs, inst.configureBtrfs, false},
		{PhaseBaseInstall, inst.installBase, false},
		{PhaseSystemConfig, inst.configureSystem, false},
		{PhaseSwap, inst.configureSwap, false},
		{PhaseBootloader, inst.installBootloader, false},
		{PhaseServices, inst.enableServices, false},
		{PhaseSoftware, inst.installSoftware, false},
		{PhaseDesktop, inst.installDesktop, inst.cfg.Desktop == config.DesktopNone},
	}

	total := 0
	for _, p := range phases {
		if !p.skip {
			total++
		}
	}

	completed := 0
	for _, p := range phases {
		if p.skip {
			continue
		}
		inst.progress <- PhaseUpdate{
			Phase:       p.phase,
			Description: p.phase.String(),
			Percent:     float64(completed) / float64(total),
		}
		if err := p.fn(); err != nil {
			inst.progress <- PhaseUpdate{
				Phase:       p.phase,
				Description: p.phase.String(),
				Percent:     float64(completed) / float64(total),
				Err:         err,
			}
			return
		}
		completed++
	}

	inst.progress <- PhaseUpdate{
		Description: "Installation complete",
		Percent:     1.0,
		Done:        true,
	}
}

// run executes a command and sends its output as a log line.
func (inst *Installer) run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		inst.log(string(out))
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, out)
	}
	return nil
}

// log sends a log line to the progress channel.
func (inst *Installer) log(line string) {
	inst.progress <- PhaseUpdate{LogLine: line}
}
