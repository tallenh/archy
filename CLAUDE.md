# CLAUDE.md

## Project

Archy is a Go TUI application that automates Arch Linux installation. It uses Bubble Tea for the terminal UI and runs system commands via `os/exec` for the actual installation.

## Build & Test

```bash
just build    # go build -o archy ./cmd/archy
just test     # go test ./...
just vet      # go vet ./...
just check    # all three
```

## Architecture

- **`cmd/archy/main.go`** — Entry point. Checks root, detects disks/timezones, builds step models, runs Bubble Tea.
- **`internal/config`** — `InstallConfig` struct shared across all steps. `DesktopEnvironment` type with packages/display manager helpers. `BlockDevice` with NVMe-aware partition naming (`/dev/nvme0n1p1` vs `/dev/sda1`).
- **`internal/tui`** — Root `Model` handles step navigation with Enter/Esc/Ctrl+C. `StepModel` interface implemented by 14 step screens in `tui/steps/`. Conditional skip logic (passphrase step skipped when encryption disabled).
- **`internal/installer`** — `Installer.Run()` executes phases sequentially in a goroutine, reporting progress via `chan PhaseUpdate`. Chroot helpers (`chrootRun`, `chrootShell`) are methods on `Installer` for logging. All output logged to `/root/archy.log`.
- **`internal/system`** — System detection (lsblk, timedatectl, /proc/meminfo) and input validation.

## Key Conventions

- Each `arch-chroot` call mounts a fresh tmpfs on `/tmp` — commands that share temp files must run in a single `chrootShell` call.
- Passwords are set via `echo "user:pass" | chpasswd` to avoid procfs exposure.
- LUKS uses PBKDF2 (not argon2id) for GRUB compatibility.
- Passphrase piped via `cmd.Stdin` to avoid process argument exposure.
- `makepkg` runs as the user, `pacman -U` installs as root (no sudo in chroot).
- GSettings overrides via schema override files, not `gsettings` CLI (no D-Bus in chroot).
- QEMU/Proxmox detected via `systemd-detect-virt` returning `kvm` or `qemu`.

## Releasing

`just release vX.Y.Z` — pushes main, creates tag, pushes tag. GitHub Actions runs GoReleaser for linux/amd64.
