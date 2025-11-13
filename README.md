# unilin — Universal Linux Installer

A minimal cross-distribution CLI that orchestrates existing package managers (apt, dnf, pacman, zypper, apk, snap, flatpak, brew, …) and custom scripts to install, update, remove, and search software.

- CLI name: `unilin` (binary file name is `universal-linux-installer` unless you build with `-o unilin`)
- Default config dir: `~/.config/unilin` (all `*.yaml` in this dir are merged)
- Logs: `~/.config/unilin/logs/unilin.log`

## Install / Build

Prerequisites: Go 1.25+

Build from source:

```bash
git clone https://github.com/viktorprogger/universal-linux-installer.git
cd universal-linux-installer
go build -o unilin
# Binary: ./unilin
```

Or install to `$GOBIN`:

```bash
go install github.com/viktorprogger/universal-linux-installer@latest
# Binary: $GOBIN/universal-linux-installer
# Tip: you can symlink it as `unilin` if you prefer that name
```

## Usage

```bash
unilin [-v] --config /path/to/any.yaml <command> [args]
```

Commands:
- `list` — list configured packages; for custom packages shows installed version when available
- `install <name>` — install a configured package or custom package (resolves dependencies)
- `remove <name>` — remove a configured package or custom package
- `update [name]` — update one package or all configured packages
- `search <query>` — search across all configured sources that support search
- `validate` — validate merged configuration against the JSON Schema

Examples:
```bash
unilin list
unilin install neovim
unilin remove git
unilin update            # all
unilin update neovim     # one
unilin search ripgrep
unilin validate          # only validate config
```

## Configuration

Place your configuration files in the directory `~/.config/unilin/` (default). You can point `--config` to any YAML file inside that directory; `unilin` will use the parent directory.

On startup, `unilin` merges all `*.yaml` files in the config directory into one effective config. If a source name or a package name is duplicated across files (including between `packages` and `custom_packages`), the app exits with an error.

On first run, `unilin` ensures the config directory exists and creates `~/.config/unilin/sources.yaml` from embedded defaults if missing. This file is part of the merge and contains a catalog of popular package managers.

Schema support:
- The configuration schema is published at: `https://raw.githubusercontent.com/viktorprogger/universal-linux-installer/HEAD/schema/unilin.schema.json`.
- The default sources file includes `$schema` for editor validation.
- You can add the same `$schema` line to your `config.yaml` for IDE assistance.

### Runtime schema validation

- On startup, after merging all YAML files, `unilin` validates the effective config against the JSON Schema (draft-07).
- The validator uses an embedded copy of the same schema. If validation fails, `unilin` prints a list of schema errors and exits with a non-zero status.
- Errors are also logged to `~/.config/unilin/logs/unilin.log`.

Top-level keys:
- `sources`: list of package manager templates (install/remove/update/search). These are shell snippets with placeholders.
- `packages`: list of simple packages managed by a specific `source`.
- `custom_packages`: list of custom packages managed by arbitrary scripts.

Schema (see `internal/config/types.go`):

```yaml
sources:
  - type: package_manager
    name: apt
    install:
      command: "apt install -y {package_list}"
      require_root: true
    remove:
      command: "apt remove -y {package_list}"
      require_root: true
    update:
      command: "apt install --only-upgrade -y {package_list}"
      require_root: true
    search:
      command: "apt search {query}"
      require_root: false

packages:
  - name: git
    source: apt
    depends_on: []

custom_packages:
  - name: mytool
    depends_on: [git]
    get_latest_version: "curl -s https://example.com/latest"
    get_installed_version: "mytool --version | sed 's/v//'"
    # Optional: decide update need yourself (stdout: true/false/1/0/yes/no)
    # compare_versions: "[ \"$(...)\" != \"$(...)\" ] && echo true || echo false"
    download:
      command: "curl -L -o /tmp/mytool.tgz https://example.com/mytool.tgz"
      require_root: false
    remove:
      command: "rm -f /usr/local/bin/mytool"
      require_root: true
    install:
      command: "tar -C /usr/local/bin -xzf /tmp/mytool.tgz mytool"
      require_root: true
```

Notes:
- `{package_list}` is replaced with the space-separated list provided by `unilin` (typically a single name).
- `{query}` is replaced in `search` commands.
- `depends_on` is supported for both `packages` and `custom_packages`. `unilin` computes a topological order and installs dependencies first.

### Permissions: require_root

- The `require_root` flag is specified per-command (inside `install/remove/update/search` for sources, and inside `download/remove/install/...` for custom packages). If omitted, the default is `false`.
- When `require_root: true`, `unilin` elevates that step via `sudo` if not already running as root. If `sudo` authentication fails, the step fails.
- Typical privileged steps:
  - Packages (from sources): `install`, `remove`, `update`.
  - Custom packages: `download` (sometimes), `remove-before-install`, `install`.

Tips for custom packages:
- During `compare_versions`, `download`, and `install`, the environment variables `latest_version` and `installed_version` are injected into your shell scripts to avoid unbound variable errors.

## Default Sources Reference

On first run, embedded defaults are copied to `~/.config/unilin/sources.yaml`. You can copy the entries you need into your config under `sources:`. The same catalog is available in the repository at `internal/assets/default-sources.yaml`.

Included managers:
- apt (Debian/Ubuntu)
- dnf (Fedora/RHEL)
- pacman (Arch)
- zypper (openSUSE)
- apk (Alpine)
- snap (Snapcraft)
- flatpak
- brew (Linuxbrew/Homebrew)

Each entry defines commands for:
- `install`: install one or more packages
- `remove`: uninstall
- `update`: upgrade packages
- `search`: search the catalog

Example (APT):
```yaml
- type: package_manager
  name: apt
  install:
    command: "apt install -y {package_list}"
    require_root: true
  remove:
    command: "apt remove -y {package_list}"
    require_root: true
  update:
    command: "apt install --only-upgrade -y {package_list}"
    require_root: true
  search:
    command: "apt search {query}"
    require_root: false
```

## Use Cases

- Consolidate installs across distros by declaring which manager to use per package.
- Mix system packages and custom artifacts (GitHub releases, tarballs) in one config.
- Keep tools up-to-date via `update` using either:
  - direct `update` template in a `source`, or
  - custom package logic (`get_latest_version`, `get_installed_version`, `compare_versions`).
- Model dependencies between entries with `depends_on` and let `unilin` resolve order.
- Reinstalling the same set of packages onto a new system.

## Logging

- Human-readable output to stdout/stderr (with colors).
- Persistent logs are written to `~/.config/unilin/logs/unilin.log`.

## Security

- `unilin` elevates only the steps that are marked with `require_root: true` using `sudo` when necessary. Review and adapt commands and `require_root` settings to your environment.
- `unilin` executes the exact shell you configure. Prefer verified sources and checksum validation in custom scripts.

## Troubleshooting

- Run with an explicit config: `unilin --config ./myconfig.yaml list`
- Check logs: `~/.config/unilin/logs/unilin.log`
- Validate your YAML keys match the schema shown above.
