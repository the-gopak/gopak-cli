package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDefaultSourcesIfMissing(t *testing.T) {
	dir, err := os.MkdirTemp("", "gopak-assets-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// First call should create sources.yaml with embedded contents
	if err := WriteDefaultSourcesIfMissing(dir); err != nil {
		t.Fatalf("WriteDefaultSourcesIfMissing: %v", err)
	}
	p := filepath.Join(dir, "sources.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("empty sources.yaml written")
	}
	if string(b) != string(defaultSources) {
		t.Fatalf("unexpected contents written")
	}

	// If file exists, it must not overwrite
	if err := os.WriteFile(p, []byte("modified"), 0o644); err != nil {
		t.Fatalf("pre-write: %v", err)
	}
	if err := WriteDefaultSourcesIfMissing(dir); err != nil {
		t.Fatalf("second call: %v", err)
	}
	b2, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read2: %v", err)
	}
	if string(b2) != "modified" {
		t.Fatalf("existing file was overwritten")
	}
}
