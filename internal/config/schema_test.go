package config

import "testing"

func TestValidateAgainstSchema_Valid(t *testing.T) {
	cfg := Config{
		Sources: []Source{{
			Type:    "apt",
			Name:    "apt",
			Install: Command{Command: "apt install -y {package_list}"},
			Remove:  Command{Command: "apt remove -y {package_list}"},
			Update:  Command{Command: "apt upgrade -y {package_list}"},
			Search:  Command{Command: "apt search {query}"},
		}},
		Packages: []Package{{
			Name:   "curl",
			Source: "apt",
		}},
		CustomPackages: []CustomPackage{{
			Name:                "mytool",
			DependsOn:           []string{"curl"},
			GetLatestVersion:    Command{Command: "echo 1.0.0"},
			GetInstalledVersion: Command{Command: "echo 1.0.0"},
			Remove:              Command{Command: "echo remove"},
			Install:             Command{Command: "echo install"},
			Update:              Command{Command: "echo update"},
		}},
	}
	if err := ValidateAgainstSchema(cfg); err != nil {
		t.Fatalf("expected valid schema, got error: %v", err)
	}
}
