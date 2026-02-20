package config

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// tomlConfig is the raw decoded form of archy.toml.
type tomlConfig struct {
	Mode     string        `toml:"mode"`
	Device   string        `toml:"device"`
	EFISize  string        `toml:"efi_size"`
	Encrypt  *bool         `toml:"encrypt"`
	Hostname string        `toml:"hostname"`
	Timezone string        `toml:"timezone"`
	Username string        `toml:"username"`
	ZRAMSize string        `toml:"zram_size"`
	Desktop  string        `toml:"desktop"`
	Shell        string        `toml:"shell"`
	SSHD         *bool         `toml:"sshd"`
	SSHPubKeyFile string       `toml:"ssh_pubkey_file"`
	Packages     []string      `toml:"packages"`
	Dotfiles     []tomlDotfile `toml:"dotfiles"`
}

type tomlDotfile struct {
	Src  string `toml:"src"`
	Dest string `toml:"dest"`
}

// LoadFileConfig loads configuration from archy.zip or archy.toml in the
// current directory (if either exists), reads password environment variables,
// validates all provided fields, and applies the results to cfg.
// archy.zip takes precedence over archy.toml.
func LoadFileConfig(cfg *InstallConfig, disks []BlockDevice, timezones []string) error {
	if err := loadEnvVars(cfg); err != nil {
		return err
	}

	// Try archy.zip first
	if _, err := os.Stat("archy.zip"); err == nil {
		return loadFromZip(cfg, disks, timezones)
	}

	// Fall back to loose archy.toml
	if _, err := os.Stat("archy.toml"); os.IsNotExist(err) {
		return nil
	}

	var tc tomlConfig
	if _, err := toml.DecodeFile("archy.toml", &tc); err != nil {
		return fmt.Errorf("archy.toml: %w", err)
	}

	return applyTomlConfig(cfg, &tc, disks, timezones, nil)
}

func loadFromZip(cfg *InstallConfig, disks []BlockDevice, timezones []string) error {
	zr, err := zip.OpenReader("archy.zip")
	if err != nil {
		return fmt.Errorf("archy.zip: %w", err)
	}

	data, err := fs.ReadFile(zr, "archy.toml")
	if err != nil {
		zr.Close()
		return fmt.Errorf("archy.zip: %w", err)
	}

	var tc tomlConfig
	if _, err := toml.Decode(string(data), &tc); err != nil {
		zr.Close()
		return fmt.Errorf("archy.zip: archy.toml: %w", err)
	}

	cfg.BundleFS = zr

	return applyTomlConfig(cfg, &tc, disks, timezones, zr)
}

func loadEnvVars(cfg *InstallConfig) error {
	if pw := os.Getenv("ARCHY_USERPW"); pw != "" {
		if err := ValidatePassword(pw); err != nil {
			return fmt.Errorf("ARCHY_USERPW: %w", err)
		}
		cfg.UserPassword = pw
	}

	if pw := os.Getenv("ARCHY_ROOTPW"); pw != "" {
		if err := ValidatePassword(pw); err != nil {
			return fmt.Errorf("ARCHY_ROOTPW: %w", err)
		}
		cfg.RootPassword = pw
	}

	if pp := os.Getenv("ARCHY_PASSPHRASE"); pp != "" {
		if err := ValidatePassphrase(pp); err != nil {
			return fmt.Errorf("ARCHY_PASSPHRASE: %w", err)
		}
		cfg.LUKSPassphrase = pp
	}

	return nil
}

func applyTomlConfig(cfg *InstallConfig, tc *tomlConfig, disks []BlockDevice, timezones []string, bundle fs.FS) error {
	// Validate and set mode
	switch tc.Mode {
	case "", "skip", "prompt":
	default:
		return fmt.Errorf("archy.toml: invalid mode %q: must be \"skip\" or \"prompt\"", tc.Mode)
	}
	if tc.Mode != "" {
		cfg.Mode = tc.Mode
	}

	// Device
	if tc.Device != "" {
		disk, ok := findDisk(tc.Device, disks)
		if !ok {
			return fmt.Errorf("archy.toml: device %q not found (use lsblk to find available devices)", tc.Device)
		}
		cfg.Device = disk
	}

	// EFI size
	if tc.EFISize != "" {
		if err := ValidatePartitionSize(tc.EFISize); err != nil {
			return fmt.Errorf("archy.toml: efi_size: %w", err)
		}
		cfg.EFISize = tc.EFISize
	}

	// Encrypt
	if tc.Encrypt != nil {
		cfg.Encrypt = *tc.Encrypt
		cfg.EncryptSet = true
	}

	// Hostname
	if tc.Hostname != "" {
		if err := ValidateHostname(tc.Hostname); err != nil {
			return fmt.Errorf("archy.toml: hostname: %w", err)
		}
		cfg.Hostname = tc.Hostname
	}

	// Timezone
	if tc.Timezone != "" {
		if !containsTimezone(tc.Timezone, timezones) {
			return fmt.Errorf("archy.toml: timezone %q not found", tc.Timezone)
		}
		cfg.Timezone = tc.Timezone
	}

	// Username
	if tc.Username != "" {
		if err := ValidateUsername(tc.Username); err != nil {
			return fmt.Errorf("archy.toml: username: %w", err)
		}
		cfg.Username = tc.Username
	}

	// ZRAM size
	if tc.ZRAMSize != "" {
		if err := ValidateZRAMSize(tc.ZRAMSize); err != nil {
			return fmt.Errorf("archy.toml: zram_size: %w", err)
		}
		cfg.ZRAMSize = tc.ZRAMSize
	}

	// Desktop
	if tc.Desktop != "" {
		de, err := ParseDesktopEnvironment(tc.Desktop)
		if err != nil {
			return fmt.Errorf("archy.toml: %w", err)
		}
		cfg.Desktop = de
		cfg.DesktopSet = true
	}

	// Shell
	if tc.Shell != "" {
		switch strings.ToLower(tc.Shell) {
		case "bash", "zsh":
			cfg.Shell = strings.ToLower(tc.Shell)
		default:
			return fmt.Errorf("archy.toml: invalid shell %q: must be \"bash\" or \"zsh\"", tc.Shell)
		}
	}

	// SSHD
	if tc.SSHD != nil {
		cfg.SSHD = *tc.SSHD
		cfg.SSHDSet = true
	}

	// SSH public key file
	if tc.SSHPubKeyFile != "" {
		var data []byte
		var err error
		if bundle != nil {
			data, err = fs.ReadFile(bundle, tc.SSHPubKeyFile)
		} else {
			data, err = os.ReadFile(tc.SSHPubKeyFile)
		}
		if err != nil {
			return fmt.Errorf("archy.toml: ssh_pubkey_file %q: %w", tc.SSHPubKeyFile, err)
		}
		key := strings.TrimSpace(string(data))
		if err := ValidateSSHPubKey(key); err != nil {
			return fmt.Errorf("archy.toml: ssh_pubkey_file %q: %w", tc.SSHPubKeyFile, err)
		}
		cfg.SSHPubKey = key
		cfg.SSHPubKeyFromConfig = true
	}

	// Packages
	cfg.Packages = append(cfg.Packages, tc.Packages...)

	// Dotfiles
	for _, df := range tc.Dotfiles {
		if df.Src == "" {
			return fmt.Errorf("archy.toml: dotfile entry missing src")
		}
		if df.Dest == "" {
			return fmt.Errorf("archy.toml: dotfile entry missing dest")
		}
		if bundle != nil {
			if _, err := fs.Stat(bundle, df.Src); err != nil {
				return fmt.Errorf("archy.toml: dotfile src %q not found in archy.zip", df.Src)
			}
		} else {
			if _, err := os.Stat(df.Src); err != nil {
				return fmt.Errorf("archy.toml: dotfile src %q: %w", df.Src, err)
			}
		}
		cfg.Dotfiles = append(cfg.Dotfiles, Dotfile{Src: df.Src, Dest: df.Dest})
	}

	return nil
}

// ParseDesktopEnvironment converts a string to a DesktopEnvironment value.
func ParseDesktopEnvironment(s string) (DesktopEnvironment, error) {
	switch strings.ToLower(s) {
	case "none":
		return DesktopNone, nil
	case "gnome":
		return DesktopGNOME, nil
	case "gnome-minimal":
		return DesktopGNOMEMinimal, nil
	case "kde":
		return DesktopKDE, nil
	case "hyprland":
		return DesktopHyprland, nil
	default:
		return DesktopNone, fmt.Errorf("invalid desktop %q: must be one of none, gnome, gnome-minimal, kde, hyprland", s)
	}
}

func findDisk(name string, disks []BlockDevice) (BlockDevice, bool) {
	for _, d := range disks {
		if d.Name == name || d.Path() == name {
			return d, true
		}
	}
	return BlockDevice{}, false
}

func containsTimezone(tz string, zones []string) bool {
	for _, z := range zones {
		if z == tz {
			return true
		}
	}
	return false
}
