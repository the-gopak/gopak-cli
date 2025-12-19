# gopak — Universal Package Manager

A minimal cross-distribution CLI that orchestrates existing package managers (apt, dnf, pacman, zypper, apk, snap, flatpak, brew, …) and custom scripts to install, update, remove, and search software.

- CLI name: `gopak` (binary file name is `gopak-cli` unless you build with `-o gopak`)
- Default config dir: `~/.config/gopak` (all `*.yaml` in this dir are merged)
- Logs: `~/.config/gopak/logs/gopak.log`

## Install / Build

Prerequisites: Go 1.25+

Build from source:

```bash
git clone https://github.com/gopak/gopak-cli.git
cd gopak-cli
go build -o gopak
# Binary: ./gopak
```

Or install to `$GOBIN`:

```bash
go install github.com/gopak/gopak-cli@latest
# Binary: $GOBIN/gopak-cli
# Tip: you can symlink it as `gopak` if you prefer that name
```

## Usage

```bash
gopak [-v] --config /path/to/any.yaml <command> [args]
```

Commands:
- `list` — list configured packages and versions
- `install <name>` — install a configured package or custom package (resolves dependencies)
- `remove <name>` — remove a configured package or custom package
- `update [name]` — update one package or all configured packages
- `search <query>` — search across all configured sources that support search
- `validate` — validate merged configuration against the JSON Schema

Examples:
```bash
gopak list
gopak install neovim
gopak remove git
gopak update            # all
gopak update neovim     # one
gopak search ripgrep
gopak validate          # only validate config
```

## Configuration

Place your configuration files in the directory `~/.config/gopak/` (default). You can point `--config` to any YAML file inside that directory; `gopak` will use the parent directory.

On startup, `gopak` merges all `*.yaml` files in the config directory into one effective config. If a source name or a package name is duplicated across files (including between `packages` and `custom_packages`), the app exits with an error.

Schema support:
- The configuration schema is published at: `https://raw.githubusercontent.com/gopak/gopak-cli/HEAD/schema/gopak.schema.json`.
- The default sources file includes `$schema` for editor validation.
- You can add the same `$schema` line to your `config.yaml` for IDE assistance.

### Runtime schema validation

- On startup, after merging all YAML files, `gopak` validates the effective config against the JSON Schema (draft-07).
- The validator uses an embedded copy of the same schema. If validation fails, `gopak` prints a list of schema errors and exits with a non-zero status.
- Errors are also logged to `~/.config/gopak/logs/gopak.log`.

Top-level keys:
- `sources`: list of package manager templates (install/remove/update/search/pre_update/get_installed_version/get_latest_version). These are shell snippets with placeholders.
- `packages`: list of simple packages managed by a specific `source`.
- `custom_packages`: list of custom packages managed by arbitrary scripts.
- `github_release_packages`: list of packages from GitHub releases.

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
    get_installed_version:
      command: "dpkg-query -W -f='${Version}' {package}"
      require_root: false
    get_latest_version:
      command: "apt-cache policy {package} | awk '/Candidate:/ {print $2}'"
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

github_release_packages:
  - name: mygithubtool
    repo: myorg/mygithubtool
    asset_pattern: "*x86_64-unknown-linux-gnu.tar.gz"
    depends_on: [git]
    get_installed_version: "mygithubtool --version | head -n1 | awk '{print $2}' | sed 's/^v//'"
    post_install:
      command: |
        notify-send "Installed mygithubtool"
      require_root: true
    remove:
      command: "rm -f /usr/local/bin/mygithubtool"
      require_root: true
```

Notes:
- `{package_list}` is replaced with the space-separated list provided by `gopak` (typically a single name).
- `{query}` is replaced in `search` commands.
- `{package}` is replaced with a single package name when executing `get_installed_version` / `get_latest_version` commands for package managers.
- `depends_on` is supported for `packages`, `custom_packages` and `github_release_packages`. `gopak` computes a topological order and installs dependencies first.
- `pre_update` for a source is executed at most once per process for a given script (identified by hash) before any `get_latest_version` is run for that source. `gopak` does not add `sudo` around `pre_update`; if you need root-only behavior, handle it inside the script (e.g. `if [ "$(id -u)" -eq 0 ]; then ...; fi`).

### Permissions: require_root

- The `require_root` flag is specified per-command (inside `install/remove/update/search` for sources, and inside `download/remove/install/...` for custom packages). If omitted, the default is `false`.
- When `require_root: true`, `gopak` elevates that step via `sudo` if not already running as root. If `sudo` authentication fails, the step fails.
- Typical privileged steps:
  - Packages (from sources): `install`, `remove`, `update`.
  - Custom packages: `download` (sometimes), `remove-before-install`, `install`.

Tips for custom packages:
- During `compare_versions`, `download`, and `install`, the environment variables `latest_version` and `installed_version` are injected into your shell scripts to avoid unbound variable errors.

## Default Sources Reference

Default sources are embedded into the binary and are always included as part of the effective config. You can copy entries you need into your own YAML files under `sources:`. The same catalog is available in the repository at `internal/assets/default-sources.yaml`.

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
 - `pre_update`: optional hook to refresh package metadata or caches before checking latest versions (runs once per unique script)
 - `get_installed_version`: print currently installed version of a single package (used by `list` and `update`)
 - `get_latest_version`: print latest available version of a single package (used by `update`)

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
  pre_update:
    command: "if [ \"$(id -u)\" -eq 0 ]; then apt update -y; fi"
    require_root: false
  search:
    command: "apt search {query}"
    require_root: false
  get_installed_version:
    command: "dpkg-query -W -f='${Version}' {package}"
    require_root: false
  get_latest_version:
    command: "apt-cache policy {package} | awk '/Candidate:/ {print $2}'"
    require_root: false
```

## Use Cases

- Consolidate installs across distros by declaring which manager to use per package.
- Mix system packages and custom artifacts (GitHub releases, tarballs) in one config.
- Keep tools up-to-date via `update` using either:
  - direct `update` template in a `source`, or
  - custom package logic (`get_latest_version`, `get_installed_version`, `compare_versions`).
- Model dependencies between entries with `depends_on` and let `gopak` resolve order.
- Reinstalling the same set of packages onto a new system.

## Logging

- Human-readable output to stdout/stderr (with colors).
- Persistent logs are written to `~/.config/gopak/logs/gopak.log`.

## Security

- `gopak` elevates only the steps that are marked with `require_root: true` using `sudo` when necessary. Review and adapt commands and `require_root` settings to your environment.
- `gopak` executes the exact shell you configure. Prefer verified sources and checksum validation in custom scripts.

## Troubleshooting

- Run with an explicit config: `gopak --config ./myconfig.yaml list`
- Check logs: `~/.config/gopak/logs/gopak.log`
- Validate your YAML keys match the schema shown above.
