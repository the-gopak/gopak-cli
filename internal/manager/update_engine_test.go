package manager

import (
	"testing"

	"github.com/viktorprogger/universal-linux-installer/internal/config"
)

type mockRunner struct{ calls []string }

func (r *mockRunner) Run(name, step, script string, require *bool) error {
	r.calls = append(r.calls, name+":"+step)
	return nil
}
func (r *mockRunner) Close() error { return nil }

func TestUpdateSelected_Custom_NoNeed(t *testing.T) {
	cfg := config.Config{CustomPackages: []config.CustomPackage{{
		Name:                "go",
		GetInstalledVersion: config.Command{Command: "echo 1.0.0"},
		GetLatestVersion:    config.Command{Command: "echo 1.0.0"},
		Install:             config.Command{Command: "echo install"},
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
		Install:             config.Command{Command: "echo install"},
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
