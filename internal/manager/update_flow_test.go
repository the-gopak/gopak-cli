package manager

import (
    "reflect"
    "testing"

    "github.com/viktorprogger/universal-linux-installer/internal/config"
)

func TestGroupTracked(t *testing.T) {
    cfg := config.Config{
        Sources: []config.Source{{Type: "package_manager", Name: "apt"}, {Type: "package_manager", Name: "dnf"}},
        Packages: []config.Package{
            {Name: "git", Source: "apt"},
            {Name: "neovim", Source: "apt"},
            {Name: "htop", Source: "dnf"},
        },
        CustomPackages: []config.CustomPackage{
            {Name: "go"}, {Name: "windsurf"},
        },
    }
    got := groupTracked(cfg)
    want := map[string][]string{
        "apt":     {"git", "neovim"},
        "dnf":     {"htop"},
        "custom":  {"go", "windsurf"},
    }
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("groupTracked mismatch:\n got=%v\nwant=%v", got, want)
    }
}

func TestGroupSourcesOnly(t *testing.T) {
    m := New(config.Config{
        Sources: []config.Source{{Type: "package_manager", Name: "apt"}},
        Packages: []config.Package{
            {Name: "git", Source: "apt"},
            {Name: "curl", Source: "apt"},
        },
    })
    got := m.groupSourcesOnly()
    want := map[string][]string{"apt": {"git", "curl"}}
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("groupSourcesOnly mismatch: got=%v want=%v", got, want)
    }
}

func TestKindOf(t *testing.T) {
    if kindOf("custom") != "custom" { t.Fatalf("kindOf custom") }
    if kindOf("apt") != "source" { t.Fatalf("kindOf source") }
}
