# Gopak

Gopak is a command-line tool for keeping a personal software list consistent. You describe the programs you want in a small YAML file, then use the same commands to install, update, remove, and check them.

It does **not** replace your operating system's package manager. Instead, it calls tools you already use—such as APT, Pacman, Flatpak, npm, or pipx—and can also manage custom download scripts and GitHub Release assets. This is useful when setting up a new machine or maintaining tools that come from different places.

> **No Go installation is required.** Download a ready-made Gopak release for your computer.

## Install Gopak

### 1. Download the latest release

Open the [latest Gopak release](https://github.com/the-gopak/gopak-cli/releases/latest), download the archive for your operating system and CPU, and extract it. The most common choices are:

| Computer | Archive |
| --- | --- |
| Linux PC (Intel/AMD 64-bit) | `gopak_linux_amd64.tar.gz` |
| Linux ARM 64-bit | `gopak_linux_arm64.tar.gz` |
| macOS on Apple silicon | `gopak_darwin_arm64.zip` |
| macOS on Intel | `gopak_darwin_amd64.zip` |
| Windows PC (Intel/AMD 64-bit) | `gopak_windows_amd64.zip` |
| Windows on ARM | `gopak_windows_arm64.zip` |

The release also includes `checksums.txt`; optionally use it to verify the downloaded file before running it.

### 2. Put the executable on your PATH

On **Linux**, run these commands from the folder containing the downloaded archive. Replace the archive and executable names if you selected another architecture.

```sh
tar -xzf gopak_linux_amd64.tar.gz
sudo install -m 0755 gopak_linux_amd64 /usr/local/bin/gopak
gopak --version
```

On **macOS**, run these commands from the folder containing the downloaded archive. Replace the archive and executable names if needed.

```sh
unzip gopak_darwin_arm64.zip
sudo install -m 0755 gopak_darwin_arm64 /usr/local/bin/gopak
gopak --version
```

On **Windows**, extract the ZIP, rename the extracted executable to `gopak.exe` if desired, and place its folder in your user `PATH`. Then open a new PowerShell window and run:

```powershell
gopak --version
```

## Quick start

1. Create Gopak's configuration directory and a configuration file.
2. Add the software you want Gopak to manage.
3. Install it with `gopak install`.

The following Linux example manages `git` and `ripgrep` through APT:

```sh
mkdir -p ~/.config/gopak
```

Create `~/.config/gopak/config.yaml` with this content:

```yaml
$schema: https://raw.githubusercontent.com/the-gopak/gopak-cli/HEAD/schema/gopak.schema.json

packages:
  - name: git
    source: apt
  - name: ripgrep
    source: apt
```

Then check the configuration and install the packages:

```sh
gopak validate
gopak install
```

Gopak asks which uninstalled packages to install. To install every listed package without being prompted, run:

```sh
gopak install --yes
```

The `apt` source is built in, so the example does not need to define it. Choose a source that is installed and works on your computer; Gopak cannot install or configure the underlying package manager for you.

## Everyday use

All commands follow this form:

```text
gopak [--config PATH] [--verbose] <command> [arguments]
```

| Command | What it does |
| --- | --- |
| `gopak list` | Show configured packages and their detected versions. |
| `gopak install [name]` | Install one configured package, or choose from all uninstalled packages. |
| `gopak remove <name>` | Remove a configured package. |
| `gopak update [name]` | Update one package, or choose from all available updates. |
| `gopak search <query>` | Search the configured sources that support searching. |
| `gopak validate` | Check the merged configuration for errors. |
| `gopak exec -- <package> [args...]` | Update a package when needed, then run its executable. |

Examples:

```sh
gopak list
gopak install neovim
gopak install --yes
gopak remove git
gopak update
gopak update neovim
gopak update --dry-run
gopak search ripgrep
gopak validate
gopak --config ./myconfig.yaml list
gopak --verbose update neovim
gopak exec -- prettier --write .
gopak exec --no-cache -- mytool --help
```

`install` and `update` support `--dry-run` to show planned work without changing anything. They support `--yes` (or `-y`) to skip interactive confirmation.

## Configuration

Gopak reads every `.yaml` and `.yml` file in `~/.config/gopak/` and merges them into one configuration. If you use `--config /path/to/file.yaml`, it instead reads all YAML files next to that file. Duplicate source or package names are errors.

A configuration has up to four main sections:

- `sources`: instructions for a package manager.
- `packages`: ordinary packages handled by a source.
- `custom_packages`: packages installed by shell commands you provide.
- `github_release_packages`: packages downloaded from GitHub Releases.

The default sources bundled with Gopak are `apt`, `pacman`, `snap`, `flatpak`, `pipx`, `npm`, and `npx`. You only need to add a `sources` entry when you need a source that is not already bundled or want to override one.

### A package-manager package

```yaml
packages:
  - name: git
    source: apt
    depends_on: []
```

`depends_on` is optional. When present, Gopak installs dependencies before the package that needs them.

### A custom package

Use a custom package when the tool is not available through one of your package managers. Gopak runs the commands exactly as written.

```yaml
custom_packages:
  - name: mytool
    executable: mytool
    get_latest_version: "curl -fsSL https://example.com/latest-version.txt"
    get_installed_version: "mytool --version 2>&1 | grep -oE '[0-9]+\\.[0-9]+\\.[0-9]+'"
    download:
      command: "curl -fsSL -o /tmp/mytool.tar.gz https://example.com/mytool-linux-amd64.tar.gz"
      require_root: false
    install:
      command: "tar -C /usr/local/bin -xzf /tmp/mytool.tar.gz mytool"
      require_root: true
    remove:
      command: "rm -f /usr/local/bin/mytool"
      require_root: true
```

### A GitHub Release package

```yaml
github_release_packages:
  - name: mygithubtool
    repo: myorg/mygithubtool
    asset_pattern: "*x86_64-unknown-linux-gnu.tar.gz"
    executable: mygithubtool
    get_installed_version: "mygithubtool --version | head -n1 | awk '{print $2}' | sed 's/^v//'"
    post_install:
      command: "install -m 0755 \"$asset_path\" /usr/local/bin/mygithubtool"
      require_root: true
    remove:
      command: "rm -f /usr/local/bin/mygithubtool"
      require_root: true
```

`asset_pattern` selects a file from the repository's latest release. For custom and GitHub Release packages, `depends_on` is also available.

### Permissions and safety

Every executable step has a `require_root` setting. When it is `true`, Gopak uses `sudo` when necessary. Package-manager installs commonly need it; downloads usually do not.

Gopak runs configured shell commands, so review configuration files before using them—especially commands that download files, remove files, or request administrator access. For custom package scripts, the `latest_version` and `installed_version` environment variables are available during version comparison, download, and installation.

## Run a tool with `exec`

`exec` is handy for a configured command-line tool you want to update automatically before using. It checks for an update at most once every three hours by default, updates the package when needed, and then runs its executable.

Set `executable` to a command name or to a command plus fixed arguments:

```yaml
packages:
  - name: prettier
    source: npm
    executable: ["npx", "-y", "prettier"]

exec_cache_ttl: 24h
```

Now run Prettier through Gopak:

```sh
gopak exec -- prettier --write .
```

Use `--no-cache` to force an update check. `exec_cache_ttl` accepts duration values such as `24h` and `30m`; this setting does not require Go to be installed.

## Updates, logs, and troubleshooting

Gopak includes itself as a GitHub Release package. After installing from a release, run the following to update Gopak when a newer release is available:

```sh
gopak update gopak-cli
```

Gopak writes logs to `~/.config/gopak/logs/gopak.log`.

If something does not work:

```sh
gopak validate
gopak --config ./myconfig.yaml list
gopak --verbose update --dry-run
```

Also confirm that the underlying package manager exists and works on your machine, and inspect the log file for the command that failed.
