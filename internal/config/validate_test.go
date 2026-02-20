package config

import "testing"

func TestValidateHostname(t *testing.T) {
	valid := []string{"archlinux", "my-host", "A123"}
	for _, v := range valid {
		if err := ValidateHostname(v); err != nil {
			t.Errorf("ValidateHostname(%q) = %v, want nil", v, err)
		}
	}
	invalid := []string{"", "-bad", "123start", "has space", "a@b"}
	for _, v := range invalid {
		if err := ValidateHostname(v); err == nil {
			t.Errorf("ValidateHostname(%q) = nil, want error", v)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	valid := []string{"user", "_admin", "john_doe", "a1"}
	for _, v := range valid {
		if err := ValidateUsername(v); err != nil {
			t.Errorf("ValidateUsername(%q) = %v, want nil", v, err)
		}
	}
	invalid := []string{"", "User", "1user", "has space", "CAPS"}
	for _, v := range invalid {
		if err := ValidateUsername(v); err == nil {
			t.Errorf("ValidateUsername(%q) = nil, want error", v)
		}
	}
}

func TestValidatePartitionSize(t *testing.T) {
	valid := []string{"512M", "1G", "256m", "2g"}
	for _, v := range valid {
		if err := ValidatePartitionSize(v); err != nil {
			t.Errorf("ValidatePartitionSize(%q) = %v, want nil", v, err)
		}
	}
	invalid := []string{"", "512", "abc", "512MB", "1T"}
	for _, v := range invalid {
		if err := ValidatePartitionSize(v); err == nil {
			t.Errorf("ValidatePartitionSize(%q) = nil, want error", v)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("test"); err != nil {
		t.Errorf("ValidatePassword(test) = %v, want nil", err)
	}
	if err := ValidatePassword("ab"); err == nil {
		t.Error("ValidatePassword(ab) = nil, want error")
	}
	if err := ValidatePassword(""); err == nil {
		t.Error("ValidatePassword('') = nil, want error")
	}
}

func TestValidatePassphrase(t *testing.T) {
	if err := ValidatePassphrase("longpassphrase"); err != nil {
		t.Errorf("ValidatePassphrase(longpassphrase) = %v, want nil", err)
	}
	if err := ValidatePassphrase("short"); err == nil {
		t.Error("ValidatePassphrase(short) = nil, want error (< 8 chars)")
	}
}

func TestValidateZRAMSize(t *testing.T) {
	valid := []string{"8G", "4096M", "2g", "512m", "ram / 2", "ram/4"}
	for _, v := range valid {
		if err := ValidateZRAMSize(v); err != nil {
			t.Errorf("ValidateZRAMSize(%q) = %v, want nil", v, err)
		}
	}
	invalid := []string{"", "abc", "8"}
	for _, v := range invalid {
		if err := ValidateZRAMSize(v); err == nil {
			t.Errorf("ValidateZRAMSize(%q) = nil, want error", v)
		}
	}
}
