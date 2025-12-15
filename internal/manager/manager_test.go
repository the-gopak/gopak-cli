package manager

import (
	"testing"

	"github.com/gopak/gopak-cli/internal/config"
)

func TestResolveOrder(t *testing.T) {
	cfg := config.Config{
		Sources: []config.Source{{Type: "package_manager", Name: "apt", Install: config.Command{Command: "echo install {package_list}"}}},
		Packages: []config.Package{
			{Name: "git", Source: "apt"},
			{Name: "neovim", Source: "apt", DependsOn: []string{"git"}},
		},
		CustomPackages: []config.CustomPackage{
			{Name: "mytool", DependsOn: []string{"neovim"}, Install: config.Command{Command: "echo install mytool"}},
		},
		GithubReleasePackages: []config.GithubReleasePackage{
			{Name: "syncthing", Repo: "syncthing/syncthing", AssetPattern: "*linux-amd64*", DependsOn: []string{"mytool"}},
		},
	}
	m := New(cfg)
	order, err := m.resolve("syncthing")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if !(pos["git"] < pos["neovim"] && pos["neovim"] < pos["mytool"] && pos["mytool"] < pos["syncthing"]) {
		t.Fatalf("bad order: %v", order)
	}
	for i, n := range order {
		pos[n] = i
	}
	if !(pos["git"] < pos["neovim"] && pos["neovim"] < pos["mytool"] && pos["mytool"] < pos["syncthing"]) {
		t.Fatalf("bad order: %v", order)
	}
}

func TestHasCommand_GithubRelease(t *testing.T) {
	cfg := config.Config{
		GithubReleasePackages: []config.GithubReleasePackage{{
			Name:         "syncthing",
			Repo:         "syncthing/syncthing",
			AssetPattern: "syncthing-linux-amd64-*.tar.gz",
			PostInstall:  config.Command{Command: "echo install"},
		}},
	}
	m := New(cfg)
	key := PackageKey{Source: "github", Name: "syncthing", Kind: "github"}
	if !m.HasCommand(key, OpInstall) {
		t.Fatalf("expected HasCommand to be true for github install")
	}
	if !m.HasCommand(key, OpUpdate) {
		t.Fatalf("expected HasCommand to be true for github update")
	}
}

func TestCompareVersions_VPrefixEqual(t *testing.T) {
	if CompareVersions("v2.0.12", "2.0.12") != 0 {
		t.Fatalf("expected versions to be equal")
	}
}

func TestResolveUnknown(t *testing.T) {
	m := New(config.Config{})
	if _, err := m.resolve("missing"); err == nil {
		t.Fatalf("expected error for unknown package")
	}
}

func TestCmpVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.4", "1.2.3", 1},
		{"1.10.0", "1.2.9", 1},
		{"v1.2.3", "1.2.3", 0},
		{"1.2.3-beta", "1.2.3", 0},
		{"2", "10", -1},
		{"1.2", "1.2.0", 0},
		{"1.2.0", "1.2", 0},
	}
	sign := func(x int) int {
		if x > 0 {
			return 1
		} else if x < 0 {
			return -1
		}
		return 0
	}
	for _, c := range cases {
		got := sign(cmpVersion(c.a, c.b))
		if got != c.want {
			t.Fatalf("cmpVersion(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}
