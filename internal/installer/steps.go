package installer

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tallenh/archy/internal/config"
)

func (inst *Installer) prepare() error {
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
	if _, err := inst.chrootShell("ln -sf /usr/share/zoneinfo/" + inst.cfg.Timezone + " /etc/localtime"); err != nil {
		return err
	}
	if _, err := inst.chrootRun("hwclock", "--systohc"); err != nil {
		return err
	}

	inst.log("Configuring locale...")
	if _, err := inst.chrootShell("sed -i '/en_US.UTF-8/s/^#//' /etc/locale.gen"); err != nil {
		return err
	}
	if _, err := inst.chrootRun("locale-gen"); err != nil {
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
	if _, err := inst.chrootShell(fmt.Sprintf("echo 'root:%s' | chpasswd", inst.cfg.RootPassword)); err != nil {
		return err
	}

	inst.log("Creating user " + inst.cfg.Username + "...")
	if _, err := inst.chrootRun("useradd", "-m", inst.cfg.Username); err != nil {
		return err
	}
	if _, err := inst.chrootShell(fmt.Sprintf("echo '%s:%s' | chpasswd", inst.cfg.Username, inst.cfg.UserPassword)); err != nil {
		return err
	}
	if _, err := inst.chrootRun("usermod", "-aG", "wheel,audio,video,optical,storage,input", inst.cfg.Username); err != nil {
		return err
	}

	inst.log("Configuring sudoers...")
	if _, err := inst.chrootShell("sed -i 's/^# %wheel ALL=(ALL:ALL) ALL/%wheel ALL=(ALL:ALL) ALL/' /etc/sudoers"); err != nil {
		return err
	}

	if inst.cfg.Shell == "zsh" {
		inst.log("Installing and setting zsh as default shell...")
		if _, err := inst.chrootRun("pacman", "-S", "--noconfirm", "zsh"); err != nil {
			return err
		}
		if _, err := inst.chrootRun("chsh", "-s", "/bin/zsh", inst.cfg.Username); err != nil {
			return err
		}
	}

	return nil
}

func (inst *Installer) configureSwap() error {
	inst.log("Installing zram-generator...")
	if _, err := inst.chrootRun("pacman", "-S", "--noconfirm", "zram-generator"); err != nil {
		return err
	}

	inst.log("Writing zram-generator config...")
	conf := fmt.Sprintf("[zram0]\nzram-size = %s\ncompression-algorithm = zstd\nswap-priority = 100\nfs-type = swap\n", inst.cfg.ZRAMSize)
	return os.WriteFile("/mnt/etc/systemd/zram-generator.conf", []byte(conf), 0o644)
}

func (inst *Installer) installBootloader() error {
	inst.log("Installing GRUB and efibootmgr...")
	if _, err := inst.chrootRun("pacman", "-S", "--noconfirm", "grub", "efibootmgr"); err != nil {
		return err
	}

	if inst.cfg.Encrypt {
		if err := inst.configureLUKSGrub(); err != nil {
			return err
		}
	}

	inst.log("Installing GRUB to EFI...")
	if _, err := inst.chrootRun("grub-install", "--target=x86_64-efi", "--efi-directory=/boot", "--bootloader-id=GRUB"); err != nil {
		return err
	}

	inst.log("Generating GRUB config...")
	_, err := inst.chrootRun("grub-mkconfig", "-o", "/boot/grub/grub.cfg")
	return err
}

func (inst *Installer) enableServices() error {
	inst.log("Installing and enabling NetworkManager...")
	if _, err := inst.chrootRun("pacman", "-S", "--noconfirm", "networkmanager"); err != nil {
		return err
	}
	if _, err := inst.chrootRun("systemctl", "enable", "NetworkManager"); err != nil {
		return err
	}

	// Install guest agents if running in QEMU/Proxmox
	if isQEMU() {
		inst.log("QEMU/Proxmox detected, installing guest agents...")
		if _, err := inst.chrootRun("pacman", "-S", "--noconfirm", "qemu-guest-agent", "spice-vdagent"); err != nil {
			return err
		}
		if _, err := inst.chrootRun("systemctl", "enable", "qemu-guest-agent"); err != nil {
			return err
		}
		if _, err := inst.chrootRun("systemctl", "enable", "spice-vdagentd.socket"); err != nil {
			return err
		}
	}

	return nil
}

func isQEMU() bool {
	out, err := exec.Command("systemd-detect-virt").Output()
	if err != nil {
		return false
	}
	virt := strings.TrimSpace(string(out))
	return virt == "kvm" || virt == "qemu"
}

func (inst *Installer) configureSSHD() error {
	inst.log("Installing openssh...")
	if _, err := inst.chrootRun("pacman", "-S", "--noconfirm", "openssh"); err != nil {
		return err
	}

	inst.log("Configuring sshd...")
	sshdConfig := "PermitRootLogin no\nPasswordAuthentication no\nPubkeyAuthentication yes\n"
	if err := os.MkdirAll("/mnt/etc/ssh/sshd_config.d", 0o755); err != nil {
		return fmt.Errorf("mkdir sshd_config.d: %w", err)
	}
	if err := os.WriteFile("/mnt/etc/ssh/sshd_config.d/10-archy.conf", []byte(sshdConfig), 0o644); err != nil {
		return fmt.Errorf("write sshd config: %w", err)
	}

	inst.log("Enabling sshd service...")
	if _, err := inst.chrootRun("systemctl", "enable", "sshd"); err != nil {
		return err
	}

	if inst.cfg.SSHPubKey != "" {
		inst.log("Installing SSH public key for " + inst.cfg.Username + "...")
		sshDir := fmt.Sprintf("/mnt/home/%s/.ssh", inst.cfg.Username)
		if err := os.MkdirAll(sshDir, 0o700); err != nil {
			return fmt.Errorf("mkdir .ssh: %w", err)
		}
		if err := os.WriteFile(sshDir+"/authorized_keys", []byte(inst.cfg.SSHPubKey+"\n"), 0o600); err != nil {
			return fmt.Errorf("write authorized_keys: %w", err)
		}
		chownCmd := fmt.Sprintf("chown -R %s:%s /home/%s/.ssh", inst.cfg.Username, inst.cfg.Username, inst.cfg.Username)
		if _, err := inst.chrootShell(chownCmd); err != nil {
			return fmt.Errorf("chown .ssh: %w", err)
		}
	}

	return nil
}

func (inst *Installer) installSoftware() error {
	inst.log("Installing base-devel and git...")
	if _, err := inst.chrootRun("pacman", "-S", "--noconfirm", "base-devel", "git", "go"); err != nil {
		return err
	}

	yayInstalled := false
	inst.log("Installing yay AUR helper...")
	yayCmd := fmt.Sprintf("su - %s -c 'git clone https://aur.archlinux.org/yay.git /tmp/yay && cd /tmp/yay && makepkg --noconfirm' && pacman -U --noconfirm /tmp/yay/yay-*.pkg.tar.zst", inst.cfg.Username)
	if _, err := inst.chrootShell(yayCmd); err != nil {
		// yay install is non-fatal
		inst.log("Warning: yay install failed (can be installed manually later)")
	} else {
		yayInstalled = true
	}

	if len(inst.cfg.Packages) > 0 {
		inst.log("Installing additional packages...")
		args := append([]string{"-S", "--noconfirm"}, inst.cfg.Packages...)
		if _, err := inst.chrootRun("pacman", args...); err != nil {
			return err
		}
	}

	if len(inst.cfg.AURPackages) > 0 {
		if !yayInstalled {
			inst.log("Warning: skipping AUR packages (yay not available): " +
				strings.Join(inst.cfg.AURPackages, ", "))
		} else {
			inst.log("Installing AUR packages...")
			sudoer := fmt.Sprintf("/etc/sudoers.d/90-archy-%s", inst.cfg.Username)
			nopasswd := fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL", inst.cfg.Username)
			if _, err := inst.chrootShell(fmt.Sprintf("echo '%s' > %s && chmod 440 %s", nopasswd, sudoer, sudoer)); err != nil {
				return err
			}
			yayArgs := append([]string{"-S", "--noconfirm"}, inst.cfg.AURPackages...)
			cmd := fmt.Sprintf("su - %s -c 'yay %s'",
				inst.cfg.Username, strings.Join(yayArgs, " "))
			_, yayErr := inst.chrootShell(cmd)
			if _, err := inst.chrootShell("rm -f " + sudoer); err != nil {
				inst.log("Warning: failed to remove temporary sudoers file")
			}
			if yayErr != nil {
				return yayErr
			}
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
	if _, err := inst.chrootRun("pacman", args...); err != nil {
		return err
	}

	dm := inst.cfg.Desktop.DisplayManager()
	if dm != "" {
		inst.log("Enabling " + dm + "...")
		if _, err := inst.chrootRun("systemctl", "enable", dm); err != nil {
			// Try with .service suffix
			_, err = inst.chrootRun("systemctl", "enable", dm+".service")
			if err != nil {
				return fmt.Errorf("enable %s: %w", dm, err)
			}
		}
	}

	// Apply GNOME settings overrides
	if inst.cfg.Desktop == config.DesktopGNOME || inst.cfg.Desktop == config.DesktopGNOMEMinimal {
		inst.log("Applying GNOME settings...")
		override := `[org.gnome.desktop.interface]
color-scheme='prefer-dark'

[org.gnome.desktop.background]
picture-uri=''
picture-uri-dark=''
primary-color='#231f30'
color-shading-type='solid'
`
		if err := os.WriteFile("/mnt/usr/share/glib-2.0/schemas/99-archy.gschema.override", []byte(override), 0o644); err != nil {
			return err
		}
		if _, err := inst.chrootRun("glib-compile-schemas", "/usr/share/glib-2.0/schemas/"); err != nil {
			return err
		}
	}

	return nil
}

func (inst *Installer) installDotfiles() error {
	for _, df := range inst.cfg.Dotfiles {
		dest := df.Dest
		isUserOwned := false
		if strings.HasPrefix(dest, "~/") {
			dest = "/home/" + inst.cfg.Username + dest[1:]
			isUserOwned = true
		}

		targetPath := "/mnt" + dest
		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", parentDir, err)
		}

		var (
			data []byte
			err  error
		)
		if inst.cfg.BundleFS != nil {
			data, err = fs.ReadFile(inst.cfg.BundleFS, df.Src)
		} else {
			data, err = os.ReadFile(df.Src)
		}
		if err != nil {
			return fmt.Errorf("read dotfile %s: %w", df.Src, err)
		}

		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return fmt.Errorf("write dotfile %s: %w", targetPath, err)
		}

		inst.log(fmt.Sprintf("Installed %s → %s", df.Src, df.Dest))

		if isUserOwned {
			chownCmd := fmt.Sprintf("chown %s:%s %s", inst.cfg.Username, inst.cfg.Username, dest)
			if _, err := inst.chrootShell(chownCmd); err != nil {
				return fmt.Errorf("chown %s: %w", dest, err)
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

