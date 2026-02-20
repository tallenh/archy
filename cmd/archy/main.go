package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tallenh/archy/internal/config"
	"github.com/tallenh/archy/internal/system"
	"github.com/tallenh/archy/internal/tui"
	"github.com/tallenh/archy/internal/tui/steps"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "archy must be run as root")
		os.Exit(1)
	}

	// Detect available disks
	disks, err := system.DetectDisks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect disks: %v\n", err)
		os.Exit(1)
	}
	if len(disks) == 0 {
		fmt.Fprintln(os.Stderr, "no disks found")
		os.Exit(1)
	}

	// Detect available timezones
	timezones, err := system.ListTimezones()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list timezones: %v\n", err)
		os.Exit(1)
	}

	// Default ZRAM size
	defaultZRAM := system.DefaultZRAMSize()

	cfg := &config.InstallConfig{
		EFISize:  "512M",
		ZRAMSize: defaultZRAM,
	}

	// Load config file and environment variables
	if err := config.LoadFileConfig(cfg, disks, timezones); err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	// Build step models
	stepModels := []tui.StepModel{
		steps.NewWelcome(),                        // 0
		steps.NewDevice(cfg, disks),               // 1
		steps.NewPartSize(cfg),                    // 2
		steps.NewEncrypt(cfg),                     // 3
		steps.NewPassphrase(cfg),                  // 4
		steps.NewHostname(cfg),                    // 5
		steps.NewTimezone(cfg, timezones),          // 6
		steps.NewUsername(cfg),                     // 7
		steps.NewUserPassword(cfg),                // 8
		steps.NewRootPassword(cfg),                // 9
		steps.NewZRAMSize(cfg),                    // 10
		steps.NewDesktop(cfg),                     // 11
		steps.NewShell(cfg),                       // 12
		steps.NewConfirm(cfg),                     // 13
		steps.NewInstall(cfg),                     // 14
	}

	m := tui.NewModel(cfg, stepModels)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
