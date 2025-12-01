package manager

import (
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
