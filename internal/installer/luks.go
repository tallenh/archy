package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (inst *Installer) setupLUKS() error {
	rootPart := inst.cfg.RootPartition()

	inst.log("Formatting LUKS2 partition (pbkdf2 for GRUB compatibility)...")
	cmd := exec.Command("cryptsetup", "luksFormat", "--type", "luks2", "--pbkdf", "pbkdf2", rootPart)
	cmd.Stdin = strings.NewReader(inst.cfg.LUKSPassphrase + "\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cryptsetup luksFormat: %w: %s", err, out)
	}

	inst.log("Opening LUKS device as cryptroot...")
	cmd = exec.Command("cryptsetup", "open", rootPart, "cryptroot")
	cmd.Stdin = strings.NewReader(inst.cfg.LUKSPassphrase + "\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cryptsetup open: %w: %s", err, out)
	}

	inst.log("Formatting /dev/mapper/cryptroot as btrfs...")
	return inst.run("mkfs.btrfs", "-f", "-L", "ArchRoot", "/dev/mapper/cryptroot")
}

// configureLUKSGrub sets up mkinitcpio and GRUB for LUKS-encrypted boot.
func (inst *Installer) configureLUKSGrub() error {
	rootPart := inst.cfg.RootPartition()

	// Get UUID of root partition
	inst.log("Getting UUID of encrypted partition...")
	out, err := exec.Command("blkid", "-s", "UUID", "-o", "value", rootPart).Output()
	if err != nil {
		return fmt.Errorf("blkid: %w", err)
	}
	uuid := strings.TrimSpace(string(out))

	// Update mkinitcpio.conf — add encrypt hook and btrfs to BINARIES
	inst.log("Configuring mkinitcpio for encryption...")
	mkinitPath := "/mnt/etc/mkinitcpio.conf"
	data, err := os.ReadFile(mkinitPath)
	if err != nil {
		return err
	}
	content := string(data)

	// Add btrfs to BINARIES
	content = strings.Replace(content, "BINARIES=()", "BINARIES=(btrfs)", 1)

	// Add encrypt hook before filesystems
	content = strings.Replace(content,
		"HOOKS=(base udev autodetect modconf kms keyboard keymap consolefont block filesystems fsck)",
		"HOOKS=(base udev autodetect modconf kms keyboard keymap consolefont block encrypt filesystems fsck)",
		1,
	)

	if err := os.WriteFile(mkinitPath, []byte(content), 0o644); err != nil {
		return err
	}

	// Regenerate initramfs
	inst.log("Regenerating initramfs...")
	if _, err := inst.chrootRun("mkinitcpio", "-P"); err != nil {
		return err
	}

	// Set GRUB_CMDLINE_LINUX for cryptdevice
	inst.log("Configuring GRUB for encrypted root...")
	grubDefault := "/mnt/etc/default/grub"
	grubData, err := os.ReadFile(grubDefault)
	if err != nil {
		return err
	}
	grubContent := string(grubData)
	cryptArg := fmt.Sprintf("cryptdevice=UUID=%s:cryptroot root=/dev/mapper/cryptroot", uuid)
	grubContent = strings.Replace(grubContent,
		`GRUB_CMDLINE_LINUX=""`,
		fmt.Sprintf(`GRUB_CMDLINE_LINUX="%s"`, cryptArg),
		1,
	)

	// Enable GRUB cryptodisk support
	grubContent = strings.Replace(grubContent,
		"#GRUB_ENABLE_CRYPTODISK=y",
		"GRUB_ENABLE_CRYPTODISK=y",
		1,
	)

	return os.WriteFile(grubDefault, []byte(grubContent), 0o644)
}
