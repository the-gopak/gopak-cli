package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gopak/gopak-cli/internal/config"
)

type mockRunner struct{ calls []string }

func (r *mockRunner) Run(name, step string, cmd config.Command) error {
	r.calls = append(r.calls, name+":"+step)
	return nil
}
func (r *mockRunner) Close() error { return nil }

func TestGetVersionInstalled_Custom(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "tool",
		GetInstalledVersion: config.Command{Command: "echo 1.2.3"},
	}}}
	m := New(cfg)
	k := PackageKey{Source: "custom", Name: "tool", Kind: "custom"}
	got := m.GetVersionInstalled(k)
	if got != "1.2.3" {
		t.Fatalf("installed version mismatch: got %q, want %q", got, "1.2.3")
	}
}

func TestGetVersionAvailable_Custom(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:             "tool",
		GetLatestVersion: config.Command{Command: "echo 2.0.0"},
	}}}
	m := New(cfg)
	k := PackageKey{Source: "custom", Name: "tool", Kind: "custom"}
	got := m.GetVersionAvailable(k)
	if got != "2.0.0" {
		t.Fatalf("available version mismatch: got %q, want %q", got, "2.0.0")
	}
}

func TestGetVersionAvailableDryRun_DoesNotRunPreUpdate(t *testing.T) {
	tmp := t.TempDir()
	marker := filepath.Join(tmp, "marker")

	cfg := config.Config{Sources: []config.Source{{
		Type:             "package_manager",
		Name:             "apt",
		PreUpdate:        config.Command{Command: "echo x > " + marker},
		GetLatestVersion: config.Command{Command: "echo 1.0.0"},
	}}}
	m := New(cfg)

	k := PackageKey{Source: "apt", Name: "git", Kind: "source"}
	_ = m.GetVersionAvailableDryRun(k)
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("dry-run should not execute pre_update")
	}

	_ = m.GetVersionAvailable(k)
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("GetVersionAvailable should execute pre_update")
	}
}

func TestUpdateSelected_Custom_NoNeed(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "go",
		GetInstalledVersion: config.Command{Command: "echo 1.0.0"},
		GetLatestVersion:    config.Command{Command: "echo 1.0.0"},
		Update:              config.Command{Command: "echo update"},
	}}}
	m := New(cfg)
	run := &mockRunner{}
	key := PackageKey{Source: "custom", Name: "go", Kind: "custom"}
	if err := m.UpdateSelected([]PackageKey{key}, run, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(run.calls) != 0 {
		t.Fatalf("runner should not be called when up-to-date, got %v", run.calls)
	}
}

func TestUpdateSelected_Custom_Proceed(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "tool",
		GetInstalledVersion: config.Command{Command: "echo 0.9.0"},
		GetLatestVersion:    config.Command{Command: "echo 1.0.0"},
		Update:              config.Command{Command: "echo update"},
	}}}
	m := New(cfg)
	run := &mockRunner{}
	key := PackageKey{Source: "custom", Name: "tool", Kind: "custom"}
	if err := m.UpdateSelected([]PackageKey{key}, run, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(run.calls) == 0 {
		t.Fatalf("runner should be called on update needed")
	}
}

func TestUpdateSelected_Custom_NoUpdateCommand(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "tool",
		GetInstalledVersion: config.Command{Command: "echo 0.9.0"},
		GetLatestVersion:    config.Command{Command: "echo 1.0.0"},
		Install:             config.Command{Command: "echo install"},
	}}}
	m := New(cfg)
	run := &mockRunner{}
	key := PackageKey{Source: "custom", Name: "tool", Kind: "custom"}
	if err := m.UpdateSelected([]PackageKey{key}, run, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(run.calls) != 0 {
		t.Fatalf("runner should not be called when no update command, got %v", run.calls)
	}
}

func TestUpdateSelected_Custom_NotInstalled(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "tool",
		GetInstalledVersion: config.Command{Command: "echo ''"},
		GetLatestVersion:    config.Command{Command: "echo 1.0.0"},
		Update:              config.Command{Command: "echo update"},
	}}}
	m := New(cfg)
	run := &mockRunner{}
	key := PackageKey{Source: "custom", Name: "tool", Kind: "custom"}
	if err := m.UpdateSelected([]PackageKey{key}, run, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(run.calls) != 0 {
		t.Fatalf("runner should not be called for uninstalled package, got %v", run.calls)
	}
}

func TestInstallSelected_Custom(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "tool",
		GetInstalledVersion: config.Command{Command: "echo ''"},
		GetLatestVersion:    config.Command{Command: "echo 1.0.0"},
		Install:             config.Command{Command: "echo install"},
	}}}
	m := New(cfg)
	run := &mockRunner{}
	key := PackageKey{Source: "custom", Name: "tool", Kind: "custom"}
	if err := m.InstallSelected([]PackageKey{key}, run, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(run.calls) == 0 {
		t.Fatalf("runner should be called for install")
	}
}

func TestInstallSelected_Custom_AlreadyInstalled(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "tool",
		GetInstalledVersion: config.Command{Command: "echo 1.0.0"},
		GetLatestVersion:    config.Command{Command: "echo 1.0.0"},
		Install:             config.Command{Command: "echo install"},
	}}}
	m := New(cfg)
	run := &mockRunner{}
	key := PackageKey{Source: "custom", Name: "tool", Kind: "custom"}
	if err := m.InstallSelected([]PackageKey{key}, run, nil); err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(run.calls) != 0 {
		t.Fatalf("runner should not be called for already installed package, got %v", run.calls)
	}
}

func TestHasCommand(t *testing.T) {
	cfg := config.Config{
		Sources: []config.Source{{
			Name:    "apt",
			Install: config.Command{Command: "apt install"},
			Update:  config.Command{Command: "apt upgrade"},
		}, {
			Name:    "snap",
			Install: config.Command{Command: "snap install"},
		}},
		CustomPackages: []config.CustomPackage{{
			Name:    "tool",
			Install: config.Command{Command: "echo install"},
		}, {
			Name:   "other",
			Update: config.Command{Command: "echo update"},
		}},
	}
	m := New(cfg)

	cases := []struct {
		key  PackageKey
		op   Operation
		want bool
	}{
		{PackageKey{Source: "apt", Name: "git", Kind: "source"}, OpInstall, true},
		{PackageKey{Source: "apt", Name: "git", Kind: "source"}, OpUpdate, true},
		{PackageKey{Source: "snap", Name: "code", Kind: "source"}, OpInstall, true},
		{PackageKey{Source: "snap", Name: "code", Kind: "source"}, OpUpdate, false},
		{PackageKey{Source: "custom", Name: "tool", Kind: "custom"}, OpInstall, true},
		{PackageKey{Source: "custom", Name: "tool", Kind: "custom"}, OpUpdate, false},
		{PackageKey{Source: "custom", Name: "other", Kind: "custom"}, OpInstall, false},
		{PackageKey{Source: "custom", Name: "other", Kind: "custom"}, OpUpdate, true},
		{PackageKey{Source: "unknown", Name: "x", Kind: "source"}, OpInstall, false},
	}
	for _, tc := range cases {
		got := m.HasCommand(tc.key, tc.op)
		if got != tc.want {
			t.Errorf("HasCommand(%v, %v) = %v, want %v", tc.key, tc.op, got, tc.want)
		}
	}
}
