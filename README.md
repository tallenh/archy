# Archy

A TUI installer for Arch Linux, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- UEFI boot with GRUB
- Btrfs with subvolumes (`@`, `@home`, `@snapshots`, `@var_log`)
- Optional LUKS2 disk encryption
- ZRAM swap
- Desktop environment selection: GNOME, GNOME Minimal, KDE Plasma, Hyprland, or None
- Automatic QEMU/Proxmox guest agent installation
- yay AUR helper
- Install log at `/root/archy.log`

## Usage

Download on the Arch live ISO:

```bash
curl -Lo archy https://archy.highroad.io/latest && chmod +x archy
./archy
```

Must be run as root. The wizard collects all configuration up front, then runs the install.

## Building

Requires Go 1.25+ and [just](https://github.com/casey/just).

```bash
just build    # build binary
just test     # run tests
just check    # vet + test + build
```

## Releasing

```bash
just release v0.2.0
```

Tags and pushes; GitHub Actions builds the release via GoReleaser.

## Project Structure

```
cmd/archy/main.go              Entry point, root check, system detection
internal/
  config/config.go              InstallConfig, DesktopEnvironment, BlockDevice
  system/                       Disk detection, timezone listing, validation
  tui/                          Bubble Tea wizard (14 steps)
  installer/                    Install engine, phase orchestration, LUKS support
```

## License

MIT
