package manager

import (
	"strings"
	"testing"

	"github.com/gopak/gopak-cli/internal/config"
)

func TestExpandCommandForNames_PerPackage(t *testing.T) {
	cmd := config.Command{Command: "do {package}"}
	group, expanded, err := expandCommandForNames(cmd, []string{"a", "b"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if group {
		t.Fatalf("expected group=false")
	}
	if len(expanded) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(expanded))
	}
	if expanded[0].Command != "do a" || expanded[1].Command != "do b" {
		t.Fatalf("unexpected expanded: %#v", expanded)
	}
}

func TestExpandCommandForNames_GroupPackageList(t *testing.T) {
	cmd := config.Command{Command: "do {package_list}"}
	group, expanded, err := expandCommandForNames(cmd, []string{"a", "b"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !group {
		t.Fatalf("expected group=true")
	}
	if len(expanded) != 1 {
		t.Fatalf("expected 1 command, got %d", len(expanded))
	}
	if expanded[0].Command != "do a b" {
		t.Fatalf("unexpected command: %q", expanded[0].Command)
	}
}

func TestExpandCommandForNames_GroupNoPlaceholder(t *testing.T) {
	cmd := config.Command{Command: "do something"}
	group, expanded, err := expandCommandForNames(cmd, []string{"a", "b"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !group {
		t.Fatalf("expected group=true")
	}
	if len(expanded) != 1 {
		t.Fatalf("expected 1 command, got %d", len(expanded))
	}
	if expanded[0].Command != "do something" {
		t.Fatalf("unexpected command: %q", expanded[0].Command)
	}
}

func TestExpandCommandForNames_BothPlaceholders_Error(t *testing.T) {
	cmd := config.Command{Command: "do {package} {package_list}"}
	_, _, err := expandCommandForNames(cmd, []string{"a"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "{package}") || !strings.Contains(err.Error(), "{package_list}") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpandCommandForName_UsesSingleName(t *testing.T) {
	cmd := config.Command{Command: "do {package}"}
	expanded, err := expandCommandForName(cmd, "a")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if expanded.Command != "do a" {
		t.Fatalf("unexpected command: %q", expanded.Command)
	}
}
