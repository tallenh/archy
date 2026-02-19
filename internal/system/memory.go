package system

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DefaultZRAMSize reads /proc/meminfo and returns half the total RAM as a string like "8G".
func DefaultZRAMSize() string {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return "4G"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				break
			}
			kb, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				break
			}
			gb := kb / 1024 / 1024 / 2 // half of total RAM in GB
			if gb < 1 {
				mb := kb / 1024 / 2
				return fmt.Sprintf("%dM", mb)
			}
			return fmt.Sprintf("%dG", gb)
		}
	}
	return "4G"
}
