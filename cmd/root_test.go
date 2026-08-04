package cmd

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion_PreservesLinkerVersion(t *testing.T) {
	if got := resolveVersionFromBuildInfo("v1.2.3", &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true); got != "v1.2.3" {
		t.Fatalf("resolveVersion() = %q, want linker version", got)
	}
}

func TestResolveVersion_UsesModuleVersion(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{Version: "v0.2.0"}}
	if got := resolveVersionFromBuildInfo("dev", info, true); got != "v0.2.0" {
		t.Fatalf("resolveVersion() = %q, want module version", got)
	}
}

func TestResolveVersion_DefaultsToDevWithoutUsableBuildInfo(t *testing.T) {
	cases := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
	}{
		{name: "unavailable", ok: false},
		{name: "nil", ok: true},
		{name: "empty", info: &debug.BuildInfo{Main: debug.Module{}}, ok: true},
		{name: "devel", info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, ok: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveVersionFromBuildInfo("dev", tc.info, tc.ok); got != "dev" {
				t.Fatalf("resolveVersion() = %q, want dev", got)
			}
		})
	}
}
