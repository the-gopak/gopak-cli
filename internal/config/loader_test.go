package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFiles_MergeOK(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.yaml")
	f2 := filepath.Join(dir, "b.yaml")
	os.WriteFile(f1, []byte(`
sources:
  - type: package_manager
    name: apt
    install:
      command: "i {package_list}"
    remove:
      command: "r {package_list}"
    update:
      command: "u {package_list}"
    search:
      command: "s {query}"
packages:
  - name: git
    source: apt
`), 0o644)
	os.WriteFile(f2, []byte(`
sources:
  - type: package_manager
    name: flatpak
    install:
      command: "i {package_list}"
    remove:
      command: "r {package_list}"
    update:
      command: "u {package_list}"
    search:
      command: "s {query}"
custom_packages:
  - name: mytool
    download:
      command: "echo d"
    remove:
      command: "echo rm"
    install:
      command: "echo in"
`), 0o644)
	cfg, err := LoadFromFiles([]string{f2, f1})
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if len(cfg.Sources) != 2 { t.Fatalf("want 2 sources, got %d", len(cfg.Sources)) }
	if len(cfg.Packages) != 1 { t.Fatalf("want 1 pkg, got %d", len(cfg.Packages)) }
	if len(cfg.CustomPackages) != 1 { t.Fatalf("want 1 custom pkg, got %d", len(cfg.CustomPackages)) }
}

func TestLoadFromFiles_DuplicateSource(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.yaml")
	f2 := filepath.Join(dir, "b.yaml")
	os.WriteFile(f1, []byte(`
 sources:
   - type: package_manager
     name: apt
 `), 0o644)
	os.WriteFile(f2, []byte(`
 sources:
   - type: package_manager
     name: apt
 `), 0o644)
	_, err := LoadFromFiles([]string{f1, f2})
	if err == nil { t.Fatalf("expected duplicate error") }
}

func TestLoadFromFiles_DuplicatePackageAcrossKinds(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.yaml")
	f2 := filepath.Join(dir, "b.yaml")
	os.WriteFile(f1, []byte(`
 packages:
   - name: tool
     source: apt
 `), 0o644)
	os.WriteFile(f2, []byte(`
 custom_packages:
   - name: tool
     download:
       command: echo d
     remove:
       command: echo rm
     install:
       command: echo in
 `), 0o644)
	_, err := LoadFromFiles([]string{f1, f2})
	if err == nil { t.Fatalf("expected duplicate error") }
}
