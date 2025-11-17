package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Type                string  `mapstructure:"type" yaml:"type" json:"type"`
	Name                string  `mapstructure:"name" yaml:"name" json:"name"`
	Install             Command `mapstructure:"install" yaml:"install" json:"install"`
	Remove              Command `mapstructure:"remove" yaml:"remove" json:"remove"`
	Update              Command `mapstructure:"update" yaml:"update" json:"update"`
	Search              Command `mapstructure:"search" yaml:"search" json:"search"`
	Outdated            Command `mapstructure:"outdated" yaml:"outdated" json:"outdated"`
	GetInstalledVersion Command `mapstructure:"get_installed_version" yaml:"get_installed_version" json:"get_installed_version"`
	GetLatestVersion    Command `mapstructure:"get_latest_version" yaml:"get_latest_version" json:"get_latest_version"`
}

type Package struct {
	Name      string   `mapstructure:"name" yaml:"name" json:"name"`
	Source    string   `mapstructure:"source" yaml:"source" json:"source"`
	DependsOn []string `mapstructure:"depends_on" yaml:"depends_on" json:"depends_on,omitempty"`
}

type CustomPackage struct {
	Name                string   `mapstructure:"name" yaml:"name" json:"name"`
	DependsOn           []string `mapstructure:"depends_on" yaml:"depends_on" json:"depends_on,omitempty"`
	GetLatestVersion    Command  `mapstructure:"get_latest_version" yaml:"get_latest_version" json:"get_latest_version"`
	GetInstalledVersion Command  `mapstructure:"get_installed_version" yaml:"get_installed_version" json:"get_installed_version"`
	CompareVersions     Command  `mapstructure:"compare_versions" yaml:"compare_versions" json:"compare_versions"`
	Download            Command  `mapstructure:"download" yaml:"download" json:"download"`
	Remove              Command  `mapstructure:"remove" yaml:"remove" json:"remove"`
	Install             Command  `mapstructure:"install" yaml:"install" json:"install"`
}

type Config struct {
	Sources        []Source        `mapstructure:"sources" yaml:"sources" json:"sources,omitempty"`
	Packages       []Package       `mapstructure:"packages" yaml:"packages" json:"packages,omitempty"`
	CustomPackages []CustomPackage `mapstructure:"custom_packages" yaml:"custom_packages" json:"custom_packages,omitempty"`
}

type Command struct {
	Command     string `mapstructure:"command" yaml:"command" json:"command"`
	RequireRoot *bool  `mapstructure:"require_root" yaml:"require_root" json:"require_root,omitempty"`
}

func (c *Command) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		c.Command = value.Value
		c.RequireRoot = nil
		return nil
	case yaml.MappingNode:
		var aux struct {
			Command     string `yaml:"command"`
			RequireRoot *bool  `yaml:"require_root"`
		}
		if err := value.Decode(&aux); err != nil {
			return err
		}
		c.Command = aux.Command
		c.RequireRoot = aux.RequireRoot
		return nil
	default:
		return fmt.Errorf("invalid command node kind: %d", value.Kind)
	}
}
