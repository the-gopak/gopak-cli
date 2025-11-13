package console

import (
	"strings"
	"testing"

	"github.com/viktorprogger/universal-linux-installer/internal/manager"
)

func TestRenderGroups_HideUpToDate(t *testing.T) {
	groups := map[string][]string{
		"apt":    {"fish", "nano"},
		"custom": {"go"},
	}
	status := map[manager.PackageKey]manager.VersionStatus{}
	// fish: installable only
	status[manager.PackageKey{Source: "apt", Name: "fish", Kind: "source"}] = manager.VersionStatus{Installed: "", Available: "3.0.0"}
	// nano: up-to-date
	status[manager.PackageKey{Source: "apt", Name: "nano", Kind: "source"}] = manager.VersionStatus{Installed: "2.0.0", Available: "2.0.0"}
	// go: needs update
	status[manager.PackageKey{Source: "custom", Name: "go", Kind: "custom"}] = manager.VersionStatus{Installed: "1.25.4", Available: "1.26.0"}

	outAll := renderGroups(groups, status, false)
	if !strings.Contains(outAll, "apt") || !strings.Contains(outAll, "custom") {
		t.Fatalf("groups not rendered: %q", outAll)
	}
	if !strings.Contains(outAll, "fish") || !strings.Contains(outAll, "nano") || !strings.Contains(outAll, "go") {
		t.Fatalf("packages not rendered: %q", outAll)
	}

	outHide := renderGroups(groups, status, true)
	if strings.Contains(outHide, "nano") {
		t.Fatalf("up-to-date package should be hidden when hideUpToDate=true: %q", outHide)
	}
	if !strings.Contains(outHide, "fish") || !strings.Contains(outHide, "go") {
		t.Fatalf("packages needing attention should remain: %q", outHide)
	}
}
