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
- Config file support (`archy.toml`) for pre-configured or fully automated installs
- Dotfile installation via config

## Usage

Download on the Arch live ISO:

```bash
curl -Lo archy https://archy.highroad.io/latest && chmod +x archy
./archy
```

Must be run as root. The wizard collects all configuration up front, then runs the install.

## Configuration

Archy can be pre-configured by placing an `archy.toml` file in the current directory. All fields are optional — any field not provided will be prompted interactively.

```toml
mode = "skip"  # "skip" or "prompt"

device = "sda"
efi_size = "512M"
encrypt = true
hostname = "archbox"
timezone = "America/New_York"
username = "alice"
zram_size = "ram / 2"
desktop = "gnome-minimal"
shell = "zsh"
docker = true
docker_group = true
packages = ["tmux", "neovim", "ripgrep"]

[[dotfiles]]
src = "dots/zshrc"
dest = "~/.zshrc"

[[dotfiles]]
src = "dots/tmux.conf"
dest = "~/.tmux.conf"
```

### Mode

- **`skip`** — Auto-advance past any step that has a value in the config. Steps without a value are still prompted.
- **`prompt`** — Pre-fill inputs from config but show every step, allowing the user to override.

### Fields

| Field | Example | Notes |
|-------|---------|-------|
| `device` | `"sda"` or `"/dev/nvme0n1"` | Must match a detected disk, or archy exits with an error |
| `efi_size` | `"512M"`, `"1G"` | |
| `encrypt` | `true`, `false` | |
| `hostname` | `"archbox"` | Letters, digits, hyphens; max 63 chars |
| `username` | `"alice"` | Lowercase letters, digits, `_`, `-`; max 32 chars |
| `timezone` | `"America/New_York"` | Must match `timedatectl list-timezones` |
| `zram_size` | `"8G"`, `"ram / 2"` | |
| `desktop` | `"gnome"`, `"gnome-minimal"`, `"kde"`, `"hyprland"`, `"none"` | Invalid values error with the list of valid options |
| `shell` | `"bash"`, `"zsh"` | Default shell for the user; zsh is installed automatically |
| `docker` | `true`, `false` | Install and enable Docker |
| `docker_group` | `true`, `false` | Add user to docker group (default: true) |
| `packages` | `["tmux", "neovim"]` | Additional pacman packages to install |

### Passwords

Passwords are provided via environment variables (never in the config file):

```bash
export ARCHY_USERPW="userpassword"
export ARCHY_ROOTPW="rootpassword"
export ARCHY_PASSPHRASE="lukspassphrase"
./archy
```

Password steps are skipped when the corresponding env var is set, regardless of mode. If an env var is not set, the password is prompted interactively.

### Dotfiles

The `[[dotfiles]]` section copies files into the installed system. `src` is relative to the current directory. `dest` supports `~` which expands to `/home/<username>`. Files under `~` are owned by the user; other paths are owned by root.

### Bundle

Instead of loose files, you can bundle `archy.toml` and dotfile sources into a single `archy.zip`:

```
archy.zip
├── archy.toml
├── dots/zshrc
└── dots/tmux.conf
```

Place `archy.zip` in the current directory and run archy. The zip is used as a virtual filesystem — dotfile `src` paths are read from inside the zip. If both `archy.zip` and `archy.toml` exist, the zip takes precedence.

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
  config/                       InstallConfig, config file loading, validation
  system/                       Disk detection, timezone listing, memory defaults
  tui/                          Bubble Tea wizard (14 steps)
  installer/                    Install engine, phase orchestration, LUKS, dotfiles
```

## License

MIT
