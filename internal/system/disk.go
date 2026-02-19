package system

import (
	"encoding/json"
	"os/exec"

	"github.com/tallenh/archy/internal/config"
)

type lsblkOutput struct {
	Blockdevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Type string `json:"type"`
	Model string `json:"model"`
}

// DetectDisks runs lsblk and returns a list of whole-disk block devices.
func DetectDisks() ([]config.BlockDevice, error) {
	out, err := exec.Command("lsblk", "-J", "-d", "-o", "NAME,SIZE,TYPE,MODEL").Output()
	if err != nil {
		return nil, err
	}
	var result lsblkOutput
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}
	var devices []config.BlockDevice
	for _, d := range result.Blockdevices {
		if d.Type != "disk" {
			continue
		}
		devices = append(devices, config.BlockDevice{
			Name:  d.Name,
			Size:  d.Size,
			Model: d.Model,
		})
	}
	return devices, nil
}
