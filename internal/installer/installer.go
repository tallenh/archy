package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/tallenh/archy/internal/config"
)

const LogPath = "/root/archy.log"

// Installer orchestrates the installation process.
type Installer struct {
	cfg      *config.InstallConfig
	progress chan<- PhaseUpdate
	logFile  *os.File
}

// New creates an Installer that reports progress to the given channel.
func New(cfg *config.InstallConfig, progress chan<- PhaseUpdate) *Installer {
	return &Installer{cfg: cfg, progress: progress}
}

// Run executes all installation phases in order.
func (inst *Installer) Run() {
	inst.openLog()
	defer inst.closeLog()

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
		{PhaseSSHD, inst.configureSSHD, !inst.cfg.SSHD},
		{PhaseDocker, inst.installDocker, !inst.cfg.Docker},
		{PhaseDesktop, inst.installDesktop, inst.cfg.Desktop == config.DesktopNone},
		{PhaseSoftware, inst.installSoftware, false},
		{PhaseDotfiles, inst.installDotfiles, len(inst.cfg.Dotfiles) == 0},
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
			inst.logToFile("SKIP  %s", p.phase)
			continue
		}
		inst.logToFile("START %s", p.phase)
		inst.progress <- PhaseUpdate{
			Phase:       p.phase,
			Description: p.phase.String(),
			Percent:     float64(completed) / float64(total),
		}
		if err := p.fn(); err != nil {
			inst.logToFile("FAIL  %s: %v", p.phase, err)
			inst.progress <- PhaseUpdate{
				Phase:       p.phase,
				Description: p.phase.String(),
				Percent:     float64(completed) / float64(total),
				Err:         err,
			}
			return
		}
		inst.logToFile("OK    %s", p.phase)
		completed++
	}

	// Copy log to installed system
	inst.copyLogToTarget()

	inst.progress <- PhaseUpdate{
		Description: "Installation complete",
		Percent:     1.0,
		Done:        true,
	}
}

// run executes a command and sends its output as a log line.
func (inst *Installer) run(name string, args ...string) error {
	inst.logToFile("RUN   %s %v", name, args)
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

// log sends a log line to the progress channel and writes it to the log file.
func (inst *Installer) log(line string) {
	inst.logToFile("      %s", line)
	inst.progress <- PhaseUpdate{LogLine: line}
}

func (inst *Installer) openLog() {
	f, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	inst.logFile = f
	fmt.Fprintf(f, "archy install log — %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "device=%s encrypt=%v desktop=%s\n\n", inst.cfg.Device.Path(), inst.cfg.Encrypt, inst.cfg.Desktop)
}

func (inst *Installer) closeLog() {
	if inst.logFile != nil {
		inst.logFile.Close()
	}
}

func (inst *Installer) logToFile(format string, a ...any) {
	if inst.logFile == nil {
		return
	}
	ts := time.Now().Format("15:04:05")
	fmt.Fprintf(inst.logFile, "%s %s\n", ts, fmt.Sprintf(format, a...))
}

func (inst *Installer) copyLogToTarget() {
	if inst.logFile == nil {
		return
	}
	inst.logFile.Close()

	src, err := os.Open(LogPath)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.OpenFile("/mnt/root/archy.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	defer dst.Close()

	io.Copy(dst, src)

	// Reopen for any further writes
	inst.logFile, _ = os.OpenFile(LogPath, os.O_APPEND|os.O_WRONLY, 0o644)
}
