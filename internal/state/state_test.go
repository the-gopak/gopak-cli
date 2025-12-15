package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_SetGetPackageState(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	ps := PackageState{
		Version:     "1.0.0",
		InstalledAt: time.Now().Format(time.RFC3339),
		FileChecksums: map[string]string{
			"/usr/local/bin/tool": "abc123",
		},
	}

	if err := m.SetPackageState("test-pkg", ps); err != nil {
		t.Fatalf("SetPackageState: %v", err)
	}

	got, ok := m.GetPackageState("test-pkg")
	if !ok {
		t.Fatal("GetPackageState: not found")
	}
	if got.Version != ps.Version {
		t.Errorf("Version = %q, want %q", got.Version, ps.Version)
	}
}

func TestManager_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	m1, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	ps := PackageState{Version: "2.0.0", InstalledAt: time.Now().Format(time.RFC3339)}
	if err := m1.SetPackageState("persist-pkg", ps); err != nil {
		t.Fatalf("SetPackageState: %v", err)
	}

	m2, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager (reload): %v", err)
	}

	got, ok := m2.GetPackageState("persist-pkg")
	if !ok {
		t.Fatal("GetPackageState after reload: not found")
	}
	if got.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", got.Version, "2.0.0")
	}
}

func TestFileChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	sum, err := FileChecksum(testFile)
	if err != nil {
		t.Fatalf("FileChecksum: %v", err)
	}

	if sum != "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9" {
		t.Errorf("unexpected checksum: %s", sum)
	}
}

func TestManager_VerifyChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "binary")
	content := []byte("binary content")
	if err := os.WriteFile(testFile, content, 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	sum, _ := FileChecksum(testFile)

	m, _ := NewManager(tmpDir)
	ps := PackageState{
		Version:       "1.0.0",
		InstalledAt:   time.Now().Format(time.RFC3339),
		FileChecksums: map[string]string{testFile: sum},
	}
	m.SetPackageState("test-pkg", ps)

	ok, err := m.VerifyChecksums("test-pkg", []string{testFile})
	if err != nil {
		t.Fatalf("VerifyChecksums: %v", err)
	}
	if !ok {
		t.Error("VerifyChecksums should return true for matching checksum")
	}

	os.WriteFile(testFile, []byte("modified"), 0o755)
	ok, err = m.VerifyChecksums("test-pkg", []string{testFile})
	if err != nil {
		t.Fatalf("VerifyChecksums: %v", err)
	}
	if ok {
		t.Error("VerifyChecksums should return false for modified file")
	}
}
