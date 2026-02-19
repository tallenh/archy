package installer

import (
	"fmt"
	"os"
	"os/exec"
)

func (inst *Installer) prepare() error {
	inst.log("Setting console font...")
	_ = inst.run("setfont", "ter-v32b") // non-fatal if font not available
	inst.log("Enabling NTP...")
	return inst.run("timedatectl", "set-ntp", "true")
}

func (inst *Installer) partition() error {
	dev := inst.cfg.Device.Path()
	efiPart := inst.cfg.EFIPartition()
	rootPart := inst.cfg.RootPartition()

	inst.log("Wiping partition table on " + dev + "...")
	if err := inst.run("sgdisk", "--zap-all", dev); err != nil {
		return err
	}

	inst.log("Creating EFI partition (" + inst.cfg.EFISize + ")...")
	if err := inst.run("sgdisk", "-n", "1:0:+"+inst.cfg.EFISize, "-t", "1:ef00", dev); err != nil {
		return err
	}

	inst.log("Creating root partition...")
	if err := inst.run("sgdisk", "-n", "2:0:0", "-t", "2:8300", dev); err != nil {
		return err
	}

	inst.log("Formatting EFI partition...")
	if err := inst.run("mkfs.fat", "-F32", efiPart); err != nil {
		return err
	}

	// Only format root as btrfs if not encrypting (LUKS path formats after opening)
	if !inst.cfg.Encrypt {
		inst.log("Formatting root partition as btrfs...")
		if err := inst.run("mkfs.btrfs", "-f", "-L", "ArchRoot", rootPart); err != nil {
			return err
		}
	}

	return nil
}

func (inst *Installer) configureBtrfs() error {
	btrfsDev := inst.cfg.BtrfsDevice()
	efiPart := inst.cfg.EFIPartition()
	opts := "noatime,compress=zstd"

	inst.log("Mounting btrfs root...")
	if err := inst.run("mount", btrfsDev, "/mnt"); err != nil {
		return err
	}

	subvolumes := []string{"@", "@home", "@snapshots", "@var_log"}
	for _, sv := range subvolumes {
		inst.log("Creating subvolume " + sv + "...")
		if err := inst.run("btrfs", "subvolume", "create", "/mnt/"+sv); err != nil {
			return err
		}
	}

	inst.log("Unmounting to remount with subvolumes...")
	if err := inst.run("umount", "/mnt"); err != nil {
		return err
	}

	// Mount @ subvolume
	if err := inst.run("mount", "-o", opts+",subvol=@", btrfsDev, "/mnt"); err != nil {
		return err
	}

	// Create mount points
	dirs := []string{"/mnt/boot", "/mnt/home", "/mnt/snapshots", "/mnt/var/log", "/mnt/etc"}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	// Mount remaining subvolumes
	mounts := []struct{ subvol, target string }{
		{"@home", "/mnt/home"},
		{"@snapshots", "/mnt/snapshots"},
		{"@var_log", "/mnt/var/log"},
	}
	for _, m := range mounts {
		inst.log("Mounting " + m.subvol + " at " + m.target + "...")
		if err := inst.run("mount", "-o", opts+",subvol="+m.subvol, btrfsDev, m.target); err != nil {
			return err
		}
	}

	// Mount EFI
	inst.log("Mounting EFI partition at /mnt/boot...")
	if err := inst.run("mount", efiPart, "/mnt/boot"); err != nil {
		return err
	}

	// Generate fstab
	inst.log("Generating fstab...")
	return inst.run("bash", "-c", "genfstab -U /mnt >> /mnt/etc/fstab")
}

func (inst *Installer) installBase() error {
	inst.log("Installing base system (this may take a while)...")
	return inst.run("pacstrap", "/mnt", "base", "linux", "linux-firmware", "sudo", "vim", "btrfs-progs")
}

func (inst *Installer) configureSystem() error {
	inst.log("Setting timezone to " + inst.cfg.Timezone + "...")
	if _, err := chrootShell("ln -sf /usr/share/zoneinfo/" + inst.cfg.Timezone + " /etc/localtime"); err != nil {
		return err
	}
	if _, err := chrootRun("hwclock", "--systohc"); err != nil {
		return err
	}

	inst.log("Configuring locale...")
	if _, err := chrootShell("sed -i '/en_US.UTF-8/s/^#//' /etc/locale.gen"); err != nil {
		return err
	}
	if _, err := chrootRun("locale-gen"); err != nil {
		return err
	}
	if err := os.WriteFile("/mnt/etc/locale.conf", []byte("LANG=en_US.UTF-8\n"), 0o644); err != nil {
		return err
	}

	inst.log("Setting hostname to " + inst.cfg.Hostname + "...")
	if err := os.WriteFile("/mnt/etc/hostname", []byte(inst.cfg.Hostname+"\n"), 0o644); err != nil {
		return err
	}

	inst.log("Setting root password...")
	if _, err := chrootShell(fmt.Sprintf("echo 'root:%s' | chpasswd", inst.cfg.RootPassword)); err != nil {
		return err
	}

	inst.log("Creating user " + inst.cfg.Username + "...")
	if _, err := chrootRun("useradd", "-m", inst.cfg.Username); err != nil {
		return err
	}
	if _, err := chrootShell(fmt.Sprintf("echo '%s:%s' | chpasswd", inst.cfg.Username, inst.cfg.UserPassword)); err != nil {
		return err
	}
	if _, err := chrootRun("usermod", "-aG", "wheel,audio,video,optical,storage,input", inst.cfg.Username); err != nil {
		return err
	}

	inst.log("Configuring sudoers...")
	if _, err := chrootShell("sed -i 's/^# %wheel ALL=(ALL:ALL) ALL/%wheel ALL=(ALL:ALL) ALL/' /etc/sudoers"); err != nil {
		return err
	}

	return nil
}

func (inst *Installer) configureSwap() error {
	inst.log("Installing zram-generator...")
	if _, err := chrootRun("pacman", "-S", "--noconfirm", "zram-generator"); err != nil {
		return err
	}

	inst.log("Writing zram-generator config...")
	conf := fmt.Sprintf("[zram0]\nzram-size = %s\ncompression-algorithm = zstd\nswap-priority = 100\nfs-type = swap\n", inst.cfg.ZRAMSize)
	return os.WriteFile("/mnt/etc/systemd/zram-generator.conf", []byte(conf), 0o644)
}

func (inst *Installer) installBootloader() error {
	inst.log("Installing GRUB and efibootmgr...")
	if _, err := chrootRun("pacman", "-S", "--noconfirm", "grub", "efibootmgr"); err != nil {
		return err
	}

	if inst.cfg.Encrypt {
		if err := inst.configureLUKSGrub(); err != nil {
			return err
		}
	}

	inst.log("Installing GRUB to EFI...")
	if _, err := chrootRun("grub-install", "--target=x86_64-efi", "--efi-directory=/boot", "--bootloader-id=GRUB"); err != nil {
		return err
	}

	inst.log("Generating GRUB config...")
	_, err := chrootRun("grub-mkconfig", "-o", "/boot/grub/grub.cfg")
	return err
}

func (inst *Installer) enableServices() error {
	inst.log("Installing and enabling NetworkManager...")
	if _, err := chrootRun("pacman", "-S", "--noconfirm", "networkmanager"); err != nil {
		return err
	}
	_, err := chrootRun("systemctl", "enable", "NetworkManager")
	return err
}

func (inst *Installer) installSoftware() error {
	inst.log("Installing base-devel and git...")
	if _, err := chrootRun("pacman", "-S", "--noconfirm", "base-devel", "git"); err != nil {
		return err
	}

	inst.log("Installing yay AUR helper...")
	cmds := []string{
		fmt.Sprintf("su - %s -c 'git clone https://aur.archlinux.org/yay.git /tmp/yay'", inst.cfg.Username),
		fmt.Sprintf("su - %s -c 'cd /tmp/yay && makepkg -si --noconfirm'", inst.cfg.Username),
	}
	for _, cmd := range cmds {
		if _, err := chrootShell(cmd); err != nil {
			// yay install is non-fatal
			inst.log("Warning: yay install failed (can be installed manually later)")
			return nil
		}
	}
	return nil
}

func (inst *Installer) installDesktop() error {
	pkgs := inst.cfg.Desktop.Packages()
	if len(pkgs) == 0 {
		return nil
	}

	inst.log("Installing " + inst.cfg.Desktop.String() + " packages...")
	args := append([]string{"-S", "--noconfirm"}, pkgs...)
	if _, err := chrootRun("pacman", args...); err != nil {
		return err
	}

	dm := inst.cfg.Desktop.DisplayManager()
	if dm != "" {
		inst.log("Enabling " + dm + "...")
		if _, err := chrootRun("systemctl", "enable", dm); err != nil {
			// Try with .service suffix
			_, err = chrootRun("systemctl", "enable", dm+".service")
			if err != nil {
				return fmt.Errorf("enable %s: %w", dm, err)
			}
		}
	}
	return nil
}

// cleanupMounts attempts to unmount and close LUKS. Called on failure or Ctrl+C.
func (inst *Installer) CleanupMounts() {
	targets := []string{"/mnt/boot", "/mnt/home", "/mnt/snapshots", "/mnt/var/log", "/mnt"}
	for _, t := range targets {
		_ = exec.Command("umount", "-l", t).Run()
	}
	if inst.cfg.Encrypt {
		_ = exec.Command("cryptsetup", "close", "cryptroot").Run()
	}
}

