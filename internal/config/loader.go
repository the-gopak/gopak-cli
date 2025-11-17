package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var current Config

func Get() Config { return current }

func LoadFromFiles(files []string) (Config, error) {
	combined := Config{}
	seen := map[string]string{}
	for _, f := range sortedYAML(files) {
		b, err := os.ReadFile(f)
		if err != nil {
			return Config{}, err
		}
		var part Config
		if err := yaml.Unmarshal(b, &part); err != nil {
			return Config{}, fmt.Errorf("%s: %w", f, err)
		}
		if err := checkPkgDuplicatesWithFiles(seen, part, f); err != nil {
			return Config{}, err
		}
		combined.Sources = append(combined.Sources, part.Sources...)
		combined.Packages = append(combined.Packages, part.Packages...)
		combined.CustomPackages = append(combined.CustomPackages, part.CustomPackages...)
	}
	if err := ValidateNoDuplicates(combined); err != nil {
		return Config{}, err
	}
	current = combined
	return combined, nil
}

func LoadDefaultsAndFiles(defaultsYAML []byte, files []string) (Config, error) {
	var base Config
	if len(defaultsYAML) > 0 {
		if err := yaml.Unmarshal(defaultsYAML, &base); err != nil {
			return Config{}, fmt.Errorf("defaults: %w", err)
		}
	}
	merged := base
	seen := map[string]string{}
	for _, p := range base.Packages {
		seen[p.Name] = "defaults"
	}
	for _, p := range base.CustomPackages {
		seen[p.Name] = "defaults"
	}
	for _, f := range sortedYAML(files) {
		b, err := os.ReadFile(f)
		if err != nil {
			return Config{}, err
		}
		var part Config
		if err := yaml.Unmarshal(b, &part); err != nil {
			return Config{}, fmt.Errorf("%s: %w", f, err)
		}
		if err := checkPkgDuplicatesWithFiles(seen, part, f); err != nil {
			return Config{}, err
		}
		merged = mergeConfig(merged, part)
	}
	if err := ValidateNoDuplicates(merged); err != nil {
		return Config{}, err
	}
	current = merged
	return merged, nil
}

func ValidateNoDuplicates(cfg Config) error {
	s := map[string]struct{}{}
	for _, v := range cfg.Sources {
		if _, ok := s[v.Name]; ok {
			return fmt.Errorf("duplicate source name: %s", v.Name)
		}
		s[v.Name] = struct{}{}
	}
	p := map[string]struct{}{}
	for _, v := range cfg.Packages {
		if _, ok := p[v.Name]; ok {
			return fmt.Errorf("duplicate package name: %s", v.Name)
		}
		p[v.Name] = struct{}{}
	}
	for _, v := range cfg.CustomPackages {
		if _, ok := p[v.Name]; ok {
			return fmt.Errorf("duplicate package name: %s", v.Name)
		}
		p[v.Name] = struct{}{}
	}
	return nil
}

func sortedYAML(files []string) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		lf := strings.ToLower(f)
		if strings.HasSuffix(lf, ".yaml") || strings.HasSuffix(lf, ".yml") {
			out = append(out, f)
		}
	}
	sort.Strings(out)
	return out
}

func mergeConfig(base, overlay Config) Config {
	srcs := map[string]Source{}
	for _, s := range base.Sources {
		srcs[s.Name] = s
	}
	for _, s := range overlay.Sources {
		if prev, ok := srcs[s.Name]; ok {
			srcs[s.Name] = mergeSource(prev, s)
		} else {
			srcs[s.Name] = s
		}
	}
	var sources []Source
	for _, s := range srcs {
		sources = append(sources, s)
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Name < sources[j].Name })

	packages := make([]Package, 0, len(base.Packages)+len(overlay.Packages))
	packages = append(packages, base.Packages...)
	packages = append(packages, overlay.Packages...)

	custom := make([]CustomPackage, 0, len(base.CustomPackages)+len(overlay.CustomPackages))
	custom = append(custom, base.CustomPackages...)
	custom = append(custom, overlay.CustomPackages...)

	return Config{Sources: sources, Packages: packages, CustomPackages: custom}
}

func mergeSource(a, b Source) Source {
	out := a
	if b.Type != "" {
		out.Type = b.Type
	}
	out.Install = mergeCommand(out.Install, b.Install)
	out.Remove = mergeCommand(out.Remove, b.Remove)
	out.Update = mergeCommand(out.Update, b.Update)
	out.Search = mergeCommand(out.Search, b.Search)
	out.Outdated = mergeCommand(out.Outdated, b.Outdated)
	out.GetInstalledVersion = mergeCommand(out.GetInstalledVersion, b.GetInstalledVersion)
	out.GetLatestVersion = mergeCommand(out.GetLatestVersion, b.GetLatestVersion)
	return out
}

func mergeCustomPackage(a, b CustomPackage) CustomPackage {
	out := a
	if len(b.DependsOn) > 0 {
		out.DependsOn = b.DependsOn
	}
	out.GetLatestVersion = mergeCommand(out.GetLatestVersion, b.GetLatestVersion)
	out.GetInstalledVersion = mergeCommand(out.GetInstalledVersion, b.GetInstalledVersion)
	out.CompareVersions = mergeCommand(out.CompareVersions, b.CompareVersions)
	out.Download = mergeCommand(out.Download, b.Download)
	out.Remove = mergeCommand(out.Remove, b.Remove)
	out.Install = mergeCommand(out.Install, b.Install)
	return out
}

func mergeCommand(a, b Command) Command {
	out := a
	if b.Command != "" {
		out.Command = b.Command
	}
	if b.RequireRoot != nil {
		out.RequireRoot = b.RequireRoot
	}
	return out
}

func checkPkgDuplicatesWithFiles(seen map[string]string, part Config, file string) error {
	local := map[string]struct{}{}
	for _, p := range part.Packages {
		if _, ok := local[p.Name]; ok {
			return fmt.Errorf("duplicate package '%s' found in %s", p.Name, file)
		}
		local[p.Name] = struct{}{}
	}
	for _, p := range part.CustomPackages {
		if _, ok := local[p.Name]; ok {
			return fmt.Errorf("duplicate package '%s' found in %s", p.Name, file)
		}
		local[p.Name] = struct{}{}
	}
	for _, p := range part.Packages {
		if prev, ok := seen[p.Name]; ok {
			return fmt.Errorf("duplicate package '%s' found in %s and %s", p.Name, prev, file)
		}
	}
	for _, p := range part.CustomPackages {
		if prev, ok := seen[p.Name]; ok {
			return fmt.Errorf("duplicate package '%s' found in %s and %s", p.Name, prev, file)
		}
	}
	for _, p := range part.Packages {
		seen[p.Name] = file
	}
	for _, p := range part.CustomPackages {
		seen[p.Name] = file
	}
	return nil
}
