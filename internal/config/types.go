package config

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Executable is the binary (and optional pre-set arguments) used by "gopak exec".
// Both forms are accepted in YAML/JSON:
//
//	executable: npx                       # string — binary name only
//	executable: ["npx", "-y", "prettier"] # array  — binary + fixed args
type Executable []string

func (e Executable) IsSet() bool { return len(e) > 0 && e[0] != "" }

func (e Executable) Binary() string {
	if len(e) == 0 {
		return ""
	}
	return e[0]
}

func (e Executable) Args() []string {
	if len(e) <= 1 {
		return nil
	}
	return e[1:]
}

func (e *Executable) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*e = Executable{value.Value}
		return nil
	case yaml.SequenceNode:
		var parts []string
		if err := value.Decode(&parts); err != nil {
			return err
		}
		*e = Executable(parts)
		return nil
	default:
		return fmt.Errorf("executable must be a string or a list of strings")
	}
}

// MarshalJSON serialises as a plain string when there is only one element,
// and as an array otherwise. This keeps the JSON compact for the common case.
func (e Executable) MarshalJSON() ([]byte, error) {
	if len(e) == 1 {
		return json.Marshal(e[0])
	}
	return json.Marshal([]string(e))
}

func (e *Executable) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*e = Executable{s}
		return nil
	}
	var parts []string
	if err := json.Unmarshal(data, &parts); err != nil {
		return fmt.Errorf("executable must be a string or a list of strings")
	}
	*e = Executable(parts)
	return nil
}

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
	Name       string     `mapstructure:"name" yaml:"name" json:"name"`
	Source     string     `mapstructure:"source" yaml:"source" json:"source"`
	DependsOn  []string   `mapstructure:"depends_on" yaml:"depends_on" json:"depends_on,omitempty"`
	Executable Executable `mapstructure:"executable" yaml:"executable" json:"executable,omitempty"`
}

type CustomPackage struct {
	Name                string     `mapstructure:"name" yaml:"name" json:"name"`
	Executable          Executable `mapstructure:"executable" yaml:"executable" json:"executable,omitempty"`
	DependsOn           []string   `mapstructure:"depends_on" yaml:"depends_on" json:"depends_on,omitempty"`
	GetInstalledVersion Command    `mapstructure:"get_installed_version" yaml:"get_installed_version" json:"get_installed_version"`
	GetLatestVersion    Command    `mapstructure:"get_latest_version" yaml:"get_latest_version" json:"get_latest_version"`
	Install             Command    `mapstructure:"install" yaml:"install" json:"install"`
	Update              Command    `mapstructure:"update" yaml:"update" json:"update"`
	Remove              Command    `mapstructure:"remove" yaml:"remove" json:"remove"`
}

type GithubReleasePackage struct {
	Name                string     `mapstructure:"name" yaml:"name" json:"name"`
	Executable          Executable `mapstructure:"executable" yaml:"executable" json:"executable,omitempty"`
	Repo                string     `mapstructure:"repo" yaml:"repo" json:"repo"`
	AssetPattern        string     `mapstructure:"asset_pattern" yaml:"asset_pattern" json:"asset_pattern"`
	GetInstalledVersion Command    `mapstructure:"get_installed_version" yaml:"get_installed_version" json:"get_installed_version"`
	PostInstall         Command    `mapstructure:"post_install" yaml:"post_install" json:"post_install"`
	Remove              Command    `mapstructure:"remove" yaml:"remove" json:"remove"`
	DependsOn           []string   `mapstructure:"depends_on" yaml:"depends_on" json:"depends_on,omitempty"`
}

type Config struct {
	Sources               []Source               `mapstructure:"sources" yaml:"sources" json:"sources,omitempty"`
	Packages              []Package              `mapstructure:"packages" yaml:"packages" json:"packages,omitempty"`
	CustomPackages        []CustomPackage        `mapstructure:"custom_packages" yaml:"custom_packages" json:"custom_packages,omitempty"`
	GithubReleasePackages []GithubReleasePackage `mapstructure:"github_release_packages" yaml:"github_release_packages" json:"github_release_packages,omitempty"`
	ExecCacheTTL          string                 `mapstructure:"exec_cache_ttl" yaml:"exec_cache_ttl" json:"exec_cache_ttl,omitempty"`
}

func (c Config) ParsedExecCacheTTL() time.Duration {
	if c.ExecCacheTTL == "" {
		return 3 * time.Hour
	}
	d, err := time.ParseDuration(c.ExecCacheTTL)
	if err != nil {
		return 3 * time.Hour
	}
	return d
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
