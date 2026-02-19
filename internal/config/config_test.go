package config

import "testing"

func TestPartitionPrefix_SATA(t *testing.T) {
	cfg := &InstallConfig{Device: BlockDevice{Name: "sda"}}
	if got := cfg.PartitionPrefix(); got != "/dev/sda" {
		t.Errorf("PartitionPrefix() = %q, want /dev/sda", got)
	}
	if got := cfg.EFIPartition(); got != "/dev/sda1" {
		t.Errorf("EFIPartition() = %q, want /dev/sda1", got)
	}
	if got := cfg.RootPartition(); got != "/dev/sda2" {
		t.Errorf("RootPartition() = %q, want /dev/sda2", got)
	}
}

func TestPartitionPrefix_NVMe(t *testing.T) {
	cfg := &InstallConfig{Device: BlockDevice{Name: "nvme0n1"}}
	if got := cfg.PartitionPrefix(); got != "/dev/nvme0n1p" {
		t.Errorf("PartitionPrefix() = %q, want /dev/nvme0n1p", got)
	}
	if got := cfg.EFIPartition(); got != "/dev/nvme0n1p1" {
		t.Errorf("EFIPartition() = %q, want /dev/nvme0n1p1", got)
	}
	if got := cfg.RootPartition(); got != "/dev/nvme0n1p2" {
		t.Errorf("RootPartition() = %q, want /dev/nvme0n1p2", got)
	}
}

func TestPartitionPrefix_MMC(t *testing.T) {
	cfg := &InstallConfig{Device: BlockDevice{Name: "mmcblk0"}}
	if got := cfg.PartitionPrefix(); got != "/dev/mmcblk0p" {
		t.Errorf("PartitionPrefix() = %q, want /dev/mmcblk0p", got)
	}
}

func TestBtrfsDevice(t *testing.T) {
	cfg := &InstallConfig{Device: BlockDevice{Name: "sda"}, Encrypt: false}
	if got := cfg.BtrfsDevice(); got != "/dev/sda2" {
		t.Errorf("BtrfsDevice() = %q, want /dev/sda2", got)
	}

	cfg.Encrypt = true
	if got := cfg.BtrfsDevice(); got != "/dev/mapper/cryptroot" {
		t.Errorf("BtrfsDevice() = %q, want /dev/mapper/cryptroot", got)
	}
}

func TestDesktopEnvironment_String(t *testing.T) {
	tests := []struct {
		de   DesktopEnvironment
		want string
	}{
		{DesktopNone, "None"},
		{DesktopGNOME, "GNOME"},
		{DesktopKDE, "KDE Plasma"},
		{DesktopHyprland, "Hyprland"},
	}
	for _, tt := range tests {
		if got := tt.de.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.de, got, tt.want)
		}
	}
}

func TestDesktopEnvironment_Packages(t *testing.T) {
	if pkgs := DesktopNone.Packages(); pkgs != nil {
		t.Errorf("DesktopNone.Packages() = %v, want nil", pkgs)
	}
	if pkgs := DesktopGNOME.Packages(); len(pkgs) == 0 {
		t.Error("DesktopGNOME.Packages() is empty")
	}
}

func TestDesktopEnvironment_DisplayManager(t *testing.T) {
	if dm := DesktopNone.DisplayManager(); dm != "" {
		t.Errorf("DesktopNone.DisplayManager() = %q, want empty", dm)
	}
	if dm := DesktopGNOME.DisplayManager(); dm != "gdm" {
		t.Errorf("DesktopGNOME.DisplayManager() = %q, want gdm", dm)
	}
	if dm := DesktopKDE.DisplayManager(); dm != "sddm" {
		t.Errorf("DesktopKDE.DisplayManager() = %q, want sddm", dm)
	}
}
