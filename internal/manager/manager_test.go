package manager

import (
	"github.com/viktorprogger/universal-linux-installer/internal/config"
	"os"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestEnsureRoot_RequireFalse_NoError(t *testing.T) {
	req := boolPtr(false)
	if err := ensureRoot("pkg", "step", req); err != nil {
		t.Fatalf("expected no error when require_root is false, got: %v", err)
	}
}

func TestEnsureRoot_DefaultOrTrue(t *testing.T) {
	euid := os.Geteuid()

	t.Run("require=nil default false", func(t *testing.T) {
		var req *bool
		err := ensureRoot("pkg", "step", req)
		if err != nil {
			t.Fatalf("expected nil error when require_root defaults to false, got %v", err)
		}
	})

	t.Run("require=true", func(t *testing.T) {
		req := boolPtr(true)
		err := ensureRoot("pkg", "step", req)
		if euid == 0 {
			if err != nil {
				t.Fatalf("expected nil error for root, got %v", err)
			}
		} else {
			if err == nil {
				t.Fatalf("expected error for non-root when require_root=true")
			}
		}
	})
}

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
	}
	m := New(cfg)
	order, err := m.resolve("mytool")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if !(pos["git"] < pos["neovim"] && pos["neovim"] < pos["mytool"]) {
		t.Fatalf("bad order: %v", order)
	}
	for i, n := range order {
		pos[n] = i
	}
	if !(pos["git"] < pos["neovim"] && pos["neovim"] < pos["mytool"]) {
		t.Fatalf("bad order: %v", order)
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
