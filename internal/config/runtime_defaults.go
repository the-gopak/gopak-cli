package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// AddRuntimeDefaults adds packages that must follow the executable currently
// running Gopak rather than a user-supplied configuration file.
func AddRuntimeDefaults(cfg Config) (Config, error) {
	pkg, ok := selfUpdatePackage(runtime.GOOS, runtime.GOARCH, executablePath())
	if !ok {
		return cfg, nil
	}
	cfg.GithubReleasePackages = append(cfg.GithubReleasePackages, pkg)
	if err := ValidateNoDuplicates(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func selfUpdatePackage(goos, goarch, executable string) (GithubReleasePackage, bool) {
	arch, ok := releaseArch(goos, goarch)
	if !ok || executable == "" {
		return GithubReleasePackage{}, false
	}

	assetPattern := fmt.Sprintf("gopak_%s_%s", goos, arch)
	if goos == "windows" {
		assetPattern += ".exe"
	}
	pkg := GithubReleasePackage{
		Name:                "gopak-cli",
		Repo:                "the-gopak/gopak-cli",
		AssetPattern:        assetPattern,
		GetInstalledVersion: Command{Command: versionCommand(goos, executable)},
	}
	if goos == "windows" {
		pkg.PostInstall = Command{Command: windowsPostInstallCommand(executable)}
	} else {
		pkg.PostInstall = Command{Command: unixPostInstallCommand(executable)}
	}
	return pkg, true
}

func releaseArch(goos, goarch string) (string, bool) {
	switch goos {
	case "linux":
		switch goarch {
		case "amd64", "386", "arm64", "riscv64":
			return goarch, true
		case "arm":
			return "arm_v7", true
		}
	case "darwin":
		if goarch == "amd64" || goarch == "arm64" {
			return goarch, true
		}
	case "windows":
		if goarch == "amd64" || goarch == "arm64" {
			return goarch, true
		}
	}
	return "", false
}

func executablePath() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
}

func versionCommand(goos, executable string) string {
	if goos == "windows" {
		return fmt.Sprintf("\"%s\" --version", strings.ReplaceAll(executable, "\"", "\"\""))
	}
	return fmt.Sprintf("%s --version", shellQuote(executable))
}

func unixPostInstallCommand(executable string) string {
	return fmt.Sprintf("target=%s; if [ -w \"$target\" ]; then install -m 0755 \"$asset_path\" \"$target\"; else sudo install -m 0755 \"$asset_path\" \"$target\"; fi", shellQuote(executable))
}

func windowsPostInstallCommand(executable string) string {
	return fmt.Sprintf("start \"\" /b powershell -NoProfile -Command \"$p=Get-Process -Id %d -ErrorAction SilentlyContinue; if ($p) { $p.WaitForExit() }; Copy-Item -Force $env:asset_path %s\"", os.Getpid(), powerShellQuote(executable))
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func powerShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
