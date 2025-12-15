package github

import "testing"

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"fd-v*-x86_64-unknown-linux-gnu.tar.gz", "fd-v10.2.0-x86_64-unknown-linux-gnu.tar.gz", true},
		{"fd-v*-x86_64-unknown-linux-gnu.tar.gz", "fd-v10.2.0-aarch64-unknown-linux-gnu.tar.gz", false},
		{"*.tar.gz", "file.tar.gz", true},
		{"*.tar.gz", "file.zip", false},
		{"bat-v*-x86_64-unknown-linux-gnu.tar.gz", "bat-v0.24.0-x86_64-unknown-linux-gnu.tar.gz", true},
		{"*linux*.deb", "package_1.0_linux_amd64.deb", true},
		{"*linux*.deb", "package_1.0_darwin_amd64.pkg", false},
		{"exact-name.zip", "exact-name.zip", true},
		{"exact-name.zip", "other-name.zip", false},
		{"?est.txt", "test.txt", true},
		{"?est.txt", "best.txt", true},
		{"?est.txt", "est.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			got := matchGlob(tt.pattern, tt.name)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}
