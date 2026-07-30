package config

import "testing"

func TestSelfUpdatePackage_SelectsReleaseForEachSupportedPlatform(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"linux", "amd64", "gopak_linux_amd64"},
		{"linux", "386", "gopak_linux_386"},
		{"linux", "arm64", "gopak_linux_arm64"},
		{"linux", "arm", "gopak_linux_arm_v7"},
		{"linux", "riscv64", "gopak_linux_riscv64"},
		{"darwin", "amd64", "gopak_darwin_amd64"},
		{"darwin", "arm64", "gopak_darwin_arm64"},
		{"windows", "amd64", "gopak_windows_amd64.exe"},
		{"windows", "arm64", "gopak_windows_arm64.exe"},
	}
	for _, tc := range cases {
		t.Run(tc.goos+"/"+tc.goarch, func(t *testing.T) {
			pkg, ok := selfUpdatePackage(tc.goos, tc.goarch, "/tmp/gopak")
			if !ok {
				t.Fatal("expected a self-update package")
			}
			if pkg.AssetPattern != tc.want {
				t.Fatalf("asset pattern: got %q, want %q", pkg.AssetPattern, tc.want)
			}
			if pkg.Repo != "the-gopak/gopak-cli" {
				t.Fatalf("repo: got %q", pkg.Repo)
			}
		})
	}
}

func TestSelfUpdatePackage_UnsupportedPlatformIsNotRegistered(t *testing.T) {
	if _, ok := selfUpdatePackage("freebsd", "amd64", "/tmp/gopak"); ok {
		t.Fatal("unexpected self-update package")
	}
}
