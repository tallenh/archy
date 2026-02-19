package installer

import (
	"fmt"
	"os/exec"
)

// chroot runs a command inside the chroot at /mnt.
func chroot(name string, args ...string) *exec.Cmd {
	chrootArgs := append([]string{"/mnt", name}, args...)
	return exec.Command("arch-chroot", chrootArgs...)
}

// chrootRun runs a command inside arch-chroot and returns combined output.
func chrootRun(name string, args ...string) (string, error) {
	cmd := chroot(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w: %s", name, err, out)
	}
	return string(out), nil
}

// chrootShell runs a shell command string inside arch-chroot.
func chrootShell(command string) (string, error) {
	return chrootRun("bash", "-c", command)
}
