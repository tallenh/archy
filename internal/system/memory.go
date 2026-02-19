package system

// DefaultZRAMSize returns the default zram-generator size expression.
func DefaultZRAMSize() string {
	return "ram / 2"
}
