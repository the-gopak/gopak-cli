package manager

import (
	"testing"

	"github.com/gopak/gopak-cli/internal/config"
)

func TestExecutableForPackage_Defaults(t *testing.T) {
	cfg := config.Config{
		Packages: []config.Package{{Name: "git", Source: "apt"}},
	}
	m := New(cfg)
	binary, args := m.executableForPackage("git")
	if binary != "git" {
		t.Fatalf("want binary %q, got %q", "git", binary)
	}
	if len(args) != 0 {
		t.Fatalf("want no args, got %v", args)
	}
}

func TestExecutableForPackage_StringForm(t *testing.T) {
	cfg := config.Config{
		Packages: []config.Package{{
			Name:       "prettier",
			Source:     "npm",
			Executable: config.Executable{"prettier"},
		}},
	}
	m := New(cfg)
	binary, args := m.executableForPackage("prettier")
	if binary != "prettier" {
		t.Fatalf("want binary %q, got %q", "prettier", binary)
	}
	if len(args) != 0 {
		t.Fatalf("want no args, got %v", args)
	}
}

func TestExecutableForPackage_ArrayForm(t *testing.T) {
	cfg := config.Config{
		Packages: []config.Package{{
			Name:       "prettier",
			Source:     "npm",
			Executable: config.Executable{"npx", "-y", "prettier"},
		}},
	}
	m := New(cfg)
	binary, args := m.executableForPackage("prettier")
	if binary != "npx" {
		t.Fatalf("want binary %q, got %q", "npx", binary)
	}
	wantArgs := []string{"-y", "prettier"}
	if len(args) != len(wantArgs) {
		t.Fatalf("want args %v, got %v", wantArgs, args)
	}
	for i, a := range wantArgs {
		if args[i] != a {
			t.Fatalf("args[%d]: want %q, got %q", i, a, args[i])
		}
	}
}

func TestExecutableForPackage_CustomPackage(t *testing.T) {
	cfg := config.Config{
		CustomPackages: []config.CustomPackage{{
			Name:       "mytool",
			Executable: config.Executable{"mytool"},
		}},
	}
	m := New(cfg)
	binary, args := m.executableForPackage("mytool")
	if binary != "mytool" {
		t.Fatalf("want binary %q, got %q", "mytool", binary)
	}
	if len(args) != 0 {
		t.Fatalf("want no args, got %v", args)
	}
}

func TestExecutableForPackage_CustomPackage_ArrayForm(t *testing.T) {
	cfg := config.Config{
		CustomPackages: []config.CustomPackage{{
			Name:       "mytool",
			Executable: config.Executable{"wrapper", "--flag", "mytool"},
		}},
	}
	m := New(cfg)
	binary, args := m.executableForPackage("mytool")
	if binary != "wrapper" {
		t.Fatalf("want binary %q, got %q", "wrapper", binary)
	}
	wantArgs := []string{"--flag", "mytool"}
	if len(args) != len(wantArgs) {
		t.Fatalf("want args %v, got %v", wantArgs, args)
	}
}

func TestExecutableForPackage_UnknownFallsBackToName(t *testing.T) {
	m := New(config.Config{})
	binary, args := m.executableForPackage("anytool")
	if binary != "anytool" {
		t.Fatalf("want binary %q, got %q", "anytool", binary)
	}
	if len(args) != 0 {
		t.Fatalf("want no args, got %v", args)
	}
}
