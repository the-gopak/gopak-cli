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
			CompareVersions:     Command{Command: "[ \"$latest_version\" = \"$installed_version\" ] && echo false || echo true"},
			Download:            Command{Command: "echo download"},
			Remove:              Command{Command: "echo remove"},
			Install:             Command{Command: "echo install"},
		}},
	}
	if err := ValidateAgainstSchema(cfg); err != nil {
		t.Fatalf("expected valid schema, got error: %v", err)
	}
}
