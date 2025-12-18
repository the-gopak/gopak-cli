package console

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
)

func TestRenderGroups_HideUpToDate(t *testing.T) {
	groups := map[string][]string{
		"apt":    {"fish", "nano"},
		"custom": {"go"},
	}
	status := map[manager.PackageKey]manager.VersionStatus{}
	// fish: installable only
	status[manager.PackageKey{Source: "apt", Name: "fish", Kind: "source"}] = manager.VersionStatus{Installed: "", Available: "3.0.0"}
	// nano: up-to-date
	status[manager.PackageKey{Source: "apt", Name: "nano", Kind: "source"}] = manager.VersionStatus{Installed: "2.0.0", Available: "2.0.0"}
	// go: needs update
	status[manager.PackageKey{Source: "custom", Name: "go", Kind: "custom"}] = manager.VersionStatus{Installed: "1.25.4", Available: "1.26.0"}

	outAll := renderGroups(groups, status, false)
	if !strings.Contains(outAll, "apt") || !strings.Contains(outAll, "custom") {
		t.Fatalf("groups not rendered: %q", outAll)
	}
	if !strings.Contains(outAll, "fish") || !strings.Contains(outAll, "nano") || !strings.Contains(outAll, "go") {
		t.Fatalf("packages not rendered: %q", outAll)
	}

	outHide := renderGroups(groups, status, true)
	if strings.Contains(outHide, "nano") {
		t.Fatalf("up-to-date package should be hidden when hideUpToDate=true: %q", outHide)
	}
	if !strings.Contains(outHide, "fish") || !strings.Contains(outHide, "go") {
		t.Fatalf("packages needing attention should remain: %q", outHide)
	}
}

func TestConsoleUIUpdate_DryRunDoesNotExecute(t *testing.T) {
	tmp := t.TempDir()
	marker := filepath.Join(tmp, "marker")

	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "test-update",
		GetInstalledVersion: config.Command{Command: "echo 1.0.0"},
		GetLatestVersion:    config.Command{Command: "echo 2.0.0"},
		Update:              config.Command{Command: fmt.Sprintf("echo updated > %q", marker)},
	}}}

	m := manager.New(cfg)
	ui := NewConsoleUI(m)

	if err := ui.Update("", true, true); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("marker should not be created in dry-run")
	}
}

func TestConsoleUIUpdate_ForceExecutes(t *testing.T) {
	tmp := t.TempDir()
	marker := filepath.Join(tmp, "marker")

	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "test-update",
		GetInstalledVersion: config.Command{Command: "echo 1.0.0"},
		GetLatestVersion:    config.Command{Command: "echo 2.0.0"},
		Update:              config.Command{Command: fmt.Sprintf("echo updated > %q", marker)},
	}}}

	m := manager.New(cfg)
	ui := NewConsoleUI(m)

	if err := ui.Update("", false, true); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("marker should be created in force mode")
	}
}

func TestConsoleUIInstall_DryRunDoesNotExecute(t *testing.T) {
	tmp := t.TempDir()
	marker := filepath.Join(tmp, "marker")

	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "test-install",
		GetInstalledVersion: config.Command{Command: ""},
		GetLatestVersion:    config.Command{Command: "echo 2.0.0"},
		Install:             config.Command{Command: fmt.Sprintf("echo installed > %q", marker)},
	}}}

	m := manager.New(cfg)
	ui := NewConsoleUI(m)

	if err := ui.Install("", true, true); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("marker should not be created in dry-run")
	}
}

func TestConsoleUIInstall_ForceExecutes(t *testing.T) {
	tmp := t.TempDir()
	marker := filepath.Join(tmp, "marker")

	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "test-install",
		GetInstalledVersion: config.Command{Command: ""},
		GetLatestVersion:    config.Command{Command: "echo 2.0.0"},
		Install:             config.Command{Command: fmt.Sprintf("echo installed > %q", marker)},
	}}}

	m := manager.New(cfg)
	ui := NewConsoleUI(m)

	if err := ui.Install("", false, true); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("marker should be created in force mode")
	}
}
