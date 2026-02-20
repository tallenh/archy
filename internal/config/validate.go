package config

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	hostnameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{0,62}$`)
	usernameRe = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
	partSizeRe = regexp.MustCompile(`^[0-9]+[MmGg]$`)
	zramSizeRe = regexp.MustCompile(`^[0-9]+[MmGg]$`)
	zramExprRe  = regexp.MustCompile(`^ram\s*/\s*[0-9]+$`)
	sshPubKeyRe = regexp.MustCompile(`^(ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp\d+|ssh-dss|sk-ssh-ed25519@openssh\.com|sk-ecdsa-sha2-nistp256@openssh\.com)\s+[A-Za-z0-9+/=]+(\s+\S.*)?$`)
)

// ValidateHostname checks that the hostname follows RFC 952.
func ValidateHostname(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if !hostnameRe.MatchString(s) {
		return fmt.Errorf("invalid hostname: must start with a letter, contain only letters/digits/hyphens, max 63 chars")
	}
	return nil
}

// ValidateUsername checks that the username is valid for Linux.
func ValidateUsername(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if !usernameRe.MatchString(s) {
		return fmt.Errorf("invalid username: must start with a lowercase letter or _, contain only lowercase letters/digits/_/-, max 32 chars")
	}
	return nil
}

// ValidatePartitionSize checks that a partition size string is valid (e.g. "512M", "1G").
func ValidatePartitionSize(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("partition size cannot be empty")
	}
	if !partSizeRe.MatchString(s) {
		return fmt.Errorf("invalid partition size: use format like 512M or 1G")
	}
	return nil
}

// ValidateZRAMSize checks that a ZRAM size string is valid.
// Accepts explicit sizes like "8G" or zram-generator expressions like "ram / 2".
func ValidateZRAMSize(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("ZRAM size cannot be empty")
	}
	if zramSizeRe.MatchString(s) || zramExprRe.MatchString(s) {
		return nil
	}
	return fmt.Errorf("invalid ZRAM size: use format like 8G, 4096M, or ram / 2")
}

// ValidatePassword checks that a password meets minimum requirements.
func ValidatePassword(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	if len(s) < 4 {
		return fmt.Errorf("password must be at least 4 characters")
	}
	return nil
}

// ValidatePassphrase checks that a LUKS passphrase meets minimum requirements.
func ValidatePassphrase(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("passphrase cannot be empty")
	}
	if len(s) < 8 {
		return fmt.Errorf("passphrase must be at least 8 characters")
	}
	return nil
}

// ValidateSSHPubKey checks that a string looks like a valid SSH public key.
func ValidateSSHPubKey(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("SSH public key cannot be empty")
	}
	if !sshPubKeyRe.MatchString(s) {
		return fmt.Errorf("invalid SSH public key: expected format like 'ssh-ed25519 AAAA... comment'")
	}
	return nil
}
