package config

import (
	"strings"
	"testing"
)

func TestValidatePlaceholders_Source_BothPlaceholders_Error(t *testing.T) {
	cfg := Config{Sources: []Source{{
		Type:    "package_manager",
		Name:    "test",
		Install: Command{Command: "echo {package} {package_list}"},
	}}}
	if err := ValidatePlaceholders(cfg); err == nil {
		t.Fatalf("expected error")
	} else {
		if !strings.Contains(err.Error(), "{package}") || !strings.Contains(err.Error(), "{package_list}") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestValidatePlaceholders_Source_OnlyPackage_OK(t *testing.T) {
	cfg := Config{Sources: []Source{{
		Type:    "package_manager",
		Name:    "test",
		Install: Command{Command: "echo {package}"},
	}}}
	if err := ValidatePlaceholders(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
