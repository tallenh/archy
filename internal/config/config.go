package config

import (
	"fmt"
	"io/fs"
	"strings"
)

// DesktopEnvironment represents the available desktop environment choices.
type DesktopEnvironment int

const (
	DesktopNone DesktopEnvironment = iota
	DesktopGNOME
	DesktopGNOMEMinimal
	DesktopKDE
	DesktopHyprland
)

func (d DesktopEnvironment) String() string {
	switch d {
	case DesktopGNOME:
		return "GNOME"
	case DesktopGNOMEMinimal:
		return "GNOME Minimal"
	case DesktopKDE:
		return "KDE Plasma"
	case DesktopHyprland:
		return "Hyprland"
	default:
		return "None"
	}
}

// Packages returns the pacman packages for the desktop environment.
func (d DesktopEnvironment) Packages() []string {
	switch d {
	case DesktopGNOME:
		return []string{"gnome"}
	case DesktopGNOMEMinimal:
		return []string{"gnome-shell", "gnome-control-center", "gdm"}
	case DesktopKDE:
		return []string{"plasma-meta", "kde-applications-meta", "sddm"}
	case DesktopHyprland:
		return []string{"hyprland", "kitty", "wofi", "sddm"}
	default:
		return nil
	}
}

// DisplayManager returns the systemd service name for the DE's display manager.
func (d DesktopEnvironment) DisplayManager() string {
	switch d {
	case DesktopGNOME, DesktopGNOMEMinimal:
		return "gdm"
	case DesktopKDE, DesktopHyprland:
		return "sddm"
	default:
		return ""
	}
}

// BlockDevice represents a disk detected by lsblk.
type BlockDevice struct {
	Name string // e.g. "sda", "nvme0n1"
	Size string // e.g. "500G"
	Model string
}

func (d BlockDevice) Path() string {
	return "/dev/" + d.Name
}

func (d BlockDevice) String() string {
	parts := []string{d.Path(), d.Size}
	if d.Model != "" {
		parts = append(parts, d.Model)
	}
	return strings.Join(parts, "  ")
}

// Dotfile describes a file to copy into the installed system.
type Dotfile struct {
	Src  string // path relative to CWD
	Dest string // destination path; ~ expands to /home/<username>
}

// InstallConfig holds all user-selected values for the installation.
type InstallConfig struct {
	Device         BlockDevice
	EFISize        string // e.g. "512M"
	Encrypt        bool
	LUKSPassphrase string
	Hostname       string
	Timezone       string
	Username       string
	UserPassword   string
	RootPassword   string
	ZRAMSize       string // e.g. "8G"
	Desktop        DesktopEnvironment
	Shell              string // "bash" or "zsh", empty means bash
	SSHD               bool   // install and enable openssh
	SSHPubKey          string // SSH public key content (from file or interactive)
	Dotfiles           []Dotfile
	Packages           []string // additional pacman packages to install
	AURPackages        []string // additional AUR packages to install via yay
	BundleFS           fs.FS    // zip bundle filesystem, nil when using loose files
	Mode               string   // "skip", "prompt", or "" (interactive)
	EncryptSet         bool     // true when encrypt was explicitly set via config
	DesktopSet         bool     // true when desktop was explicitly set via config
	SSHDSet            bool     // true when sshd was explicitly set via config
	SSHPubKeyFromConfig bool   // true when key was loaded from config file (requires APPROVE)
}

// PartitionPrefix returns the partition device prefix (handles NVMe "p" separator).
func (c *InstallConfig) PartitionPrefix() string {
	if strings.Contains(c.Device.Name, "nvme") || strings.Contains(c.Device.Name, "mmcblk") {
		return c.Device.Path() + "p"
	}
	return c.Device.Path()
}

// EFIPartition returns the path of the EFI partition (partition 1).
func (c *InstallConfig) EFIPartition() string {
	return c.PartitionPrefix() + "1"
}

// RootPartition returns the path of the root partition (partition 2).
func (c *InstallConfig) RootPartition() string {
	return c.PartitionPrefix() + "2"
}

// BtrfsDevice returns the device to format with btrfs — either the LUKS mapper
// device or the raw root partition.
func (c *InstallConfig) BtrfsDevice() string {
	if c.Encrypt {
		return "/dev/mapper/cryptroot"
	}
	return c.RootPartition()
}

// Summary returns a human-readable summary of the configuration for the confirm screen.
func (c *InstallConfig) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Device:       %s\n", c.Device)
	fmt.Fprintf(&b, "EFI Size:     %s\n", c.EFISize)
	fmt.Fprintf(&b, "Encryption:   %v\n", c.Encrypt)
	if c.Encrypt {
		fmt.Fprintf(&b, "Passphrase:   %s\n", strings.Repeat("*", len(c.LUKSPassphrase)))
	}
	fmt.Fprintf(&b, "Hostname:     %s\n", c.Hostname)
	fmt.Fprintf(&b, "Timezone:     %s\n", c.Timezone)
	fmt.Fprintf(&b, "Username:     %s\n", c.Username)
	fmt.Fprintf(&b, "User Pass:    %s\n", strings.Repeat("*", len(c.UserPassword)))
	fmt.Fprintf(&b, "Root Pass:    %s\n", strings.Repeat("*", len(c.RootPassword)))
	fmt.Fprintf(&b, "ZRAM Size:    %s\n", c.ZRAMSize)
	fmt.Fprintf(&b, "Desktop:      %s\n", c.Desktop)
	shell := c.Shell
	if shell == "" {
		shell = "bash"
	}
	fmt.Fprintf(&b, "Shell:        %s\n", shell)
	fmt.Fprintf(&b, "SSH Server:   %v\n", c.SSHD)
	if c.SSHD && c.SSHPubKey != "" {
		key := c.SSHPubKey
		if len(key) > 40 {
			key = key[:40] + "..."
		}
		fmt.Fprintf(&b, "SSH Pub Key:  %s\n", key)
	}
	if len(c.Packages) > 0 {
		fmt.Fprintf(&b, "Packages:     %s\n", strings.Join(c.Packages, ", "))
	}
	if len(c.AURPackages) > 0 {
		fmt.Fprintf(&b, "AUR Packages: %s\n", strings.Join(c.AURPackages, ", "))
	}
	if len(c.Dotfiles) > 0 {
		fmt.Fprintf(&b, "Dotfiles:     %d file(s)\n", len(c.Dotfiles))
	}
	return b.String()
}
