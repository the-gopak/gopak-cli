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
	PreUpdate           Command `mapstructure:"pre_update" yaml:"pre_update" json:"pre_update"`
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
	GetInstalledVersion Command  `mapstructure:"get_installed_version" yaml:"get_installed_version" json:"get_installed_version"`
	GetLatestVersion    Command  `mapstructure:"get_latest_version" yaml:"get_latest_version" json:"get_latest_version"`
	Install             Command  `mapstructure:"install" yaml:"install" json:"install"`
	Update              Command  `mapstructure:"update" yaml:"update" json:"update"`
	Remove              Command  `mapstructure:"remove" yaml:"remove" json:"remove"`
}

type Config struct {
	Sources        []Source        `mapstructure:"sources" yaml:"sources" json:"sources,omitempty"`
	Packages       []Package       `mapstructure:"packages" yaml:"packages" json:"packages,omitempty"`
	CustomPackages []CustomPackage `mapstructure:"custom_packages" yaml:"custom_packages" json:"custom_packages,omitempty"`
}

type Command struct {
	Command     string `mapstructure:"command" yaml:"command" json:"command"`
	RequireRoot bool   `mapstructure:"require_root" yaml:"require_root" json:"require_root"`
}

func (c *Command) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		c.Command = value.Value
		c.RequireRoot = false
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
		if aux.RequireRoot != nil {
			c.RequireRoot = *aux.RequireRoot
		} else {
			c.RequireRoot = false
		}
		return nil
	default:
		return fmt.Errorf("invalid command node kind: %d", value.Kind)
	}
}
