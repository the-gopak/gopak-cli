package manager

import "testing"

func TestExtractDigest(t *testing.T) {
	got := extractDigest("composer@sha256:abc")
	if got != "sha256:abc" {
		t.Fatalf("digest mismatch: got %q", got)
	}
}

func TestDigestFromManifestInspect_SelectPlatform(t *testing.T) {
	raw := []byte(`[
  {
    "Descriptor": {"platform": {"architecture": "amd64", "os": "linux"}},
    "OCIManifest": {"config": {"digest": "sha256:cfg-amd64"}}
  },
  {
    "Descriptor": {"platform": {"architecture": "arm64", "os": "linux"}},
    "OCIManifest": {"config": {"digest": "sha256:cfg-arm64"}}
  }
]`)

	got, err := configDigestFromManifestInspectVerbose(raw, dockerPlatform{OS: "linux", Architecture: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "sha256:cfg-arm64" {
		t.Fatalf("digest mismatch: got %q", got)
	}
}

func TestDigestFromManifestInspectVerbose_NoMatch(t *testing.T) {
	raw := []byte(`[
  {
    "Descriptor": {"platform": {"architecture": "amd64", "os": "linux"}},
    "OCIManifest": {"config": {"digest": "sha256:cfg-amd64"}}
  }
]`)
	got, err := configDigestFromManifestInspectVerbose(raw, dockerPlatform{OS: "linux", Architecture: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("digest mismatch: got %q", got)
	}
}
