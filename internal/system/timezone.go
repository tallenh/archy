package system

import (
	"os/exec"
	"strings"
)

// ListTimezones returns all available timezones via timedatectl.
func ListTimezones() ([]string, error) {
	out, err := exec.Command("timedatectl", "list-timezones").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var zones []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			zones = append(zones, l)
		}
	}
	return zones, nil
}
