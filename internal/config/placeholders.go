package config

import "fmt"

const (
	placeholderPackage     = "{package}"
	placeholderPackageList = "{package_list}"
)

func ValidatePlaceholders(cfg Config) error {
	for _, s := range cfg.Sources {
		if err := validateCommandPlaceholders("source", s.Name, "install", s.Install); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("source", s.Name, "remove", s.Remove); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("source", s.Name, "update", s.Update); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("source", s.Name, "search", s.Search); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("source", s.Name, "pre_update", s.PreUpdate); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("source", s.Name, "get_installed_version", s.GetInstalledVersion); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("source", s.Name, "get_latest_version", s.GetLatestVersion); err != nil {
			return err
		}
	}

	for _, cp := range cfg.CustomPackages {
		if err := validateCommandPlaceholders("custom_package", cp.Name, "get_installed_version", cp.GetInstalledVersion); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("custom_package", cp.Name, "get_latest_version", cp.GetLatestVersion); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("custom_package", cp.Name, "install", cp.Install); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("custom_package", cp.Name, "update", cp.Update); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("custom_package", cp.Name, "remove", cp.Remove); err != nil {
			return err
		}
	}

	for _, gp := range cfg.GithubReleasePackages {
		if err := validateCommandPlaceholders("github_release_package", gp.Name, "get_installed_version", gp.GetInstalledVersion); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("github_release_package", gp.Name, "post_install", gp.PostInstall); err != nil {
			return err
		}
		if err := validateCommandPlaceholders("github_release_package", gp.Name, "remove", gp.Remove); err != nil {
			return err
		}
	}

	return nil
}

func validateCommandPlaceholders(kind, name, field string, cmd Command) error {
	if cmd.Command == "" {
		return nil
	}
	hasPkg := contains(cmd.Command, placeholderPackage)
	hasList := contains(cmd.Command, placeholderPackageList)
	if hasPkg && hasList {
		return fmt.Errorf("invalid placeholders: %s %q %s command contains both %s and %s", kind, name, field, placeholderPackage, placeholderPackageList)
	}
	return nil
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
