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
    for _, f := range sortedYAML(files) {
        b, err := os.ReadFile(f)
        if err != nil { return Config{}, err }
        var part Config
        if err := yaml.Unmarshal(b, &part); err != nil { return Config{}, fmt.Errorf("%s: %w", f, err) }
        combined.Sources = append(combined.Sources, part.Sources...)
        combined.Packages = append(combined.Packages, part.Packages...)
        combined.CustomPackages = append(combined.CustomPackages, part.CustomPackages...)
    }
    if err := ValidateNoDuplicates(combined); err != nil { return Config{}, err }
    current = combined
    return combined, nil
}

func ValidateNoDuplicates(cfg Config) error {
    s := map[string]struct{}{}
    for _, v := range cfg.Sources {
        if _, ok := s[v.Name]; ok { return fmt.Errorf("duplicate source name: %s", v.Name) }
        s[v.Name] = struct{}{}
    }
    p := map[string]struct{}{}
    for _, v := range cfg.Packages {
        if _, ok := p[v.Name]; ok { return fmt.Errorf("duplicate package name: %s", v.Name) }
        p[v.Name] = struct{}{}
    }
    for _, v := range cfg.CustomPackages {
        if _, ok := p[v.Name]; ok { return fmt.Errorf("duplicate package name: %s", v.Name) }
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
