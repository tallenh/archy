package installer

import (
	"fmt"
	"os/exec"
)

// chroot creates a command to run inside the chroot at /mnt.
func chroot(name string, args ...string) *exec.Cmd {
	chrootArgs := append([]string{"/mnt", name}, args...)
	return exec.Command("arch-chroot", chrootArgs...)
}

// chrootRun runs a command inside arch-chroot and returns combined output.
func (inst *Installer) chrootRun(name string, args ...string) (string, error) {
	inst.logToFile("RUN   arch-chroot /mnt %s %v", name, args)
	cmd := chroot(name, args...)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		inst.logToFile("      %s", string(out))
	}
	if err != nil {
		return string(out), fmt.Errorf("%s: %w: %s", name, err, out)
	}
	return string(out), nil
}

// chrootShell runs a shell command string inside arch-chroot.
func (inst *Installer) chrootShell(command string) (string, error) {
	inst.logToFile("RUN   arch-chroot /mnt bash -c %q", command)
	cmd := chroot("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		inst.logToFile("      %s", string(out))
	}
	if err != nil {
		return string(out), fmt.Errorf("bash -c %q: %w: %s", command, err, out)
	}
	return string(out), nil
}
