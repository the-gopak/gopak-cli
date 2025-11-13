package config

import (
	"os"
	"path/filepath"
	"strings"
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
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(cfg.Sources) != 2 {
		t.Fatalf("want 2 sources, got %d", len(cfg.Sources))
	}
	if len(cfg.Packages) != 1 {
		t.Fatalf("want 1 pkg, got %d", len(cfg.Packages))
	}
	if len(cfg.CustomPackages) != 1 {
		t.Fatalf("want 1 custom pkg, got %d", len(cfg.CustomPackages))
	}
}

func TestLoadDefaultsAndFiles_DuplicateAcrossFiles_ErrorMentionsFiles(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.yaml")
	f2 := filepath.Join(dir, "b.yaml")
	os.WriteFile(f1, []byte(`
packages:
  - name: tool
    source: apt
`), 0o644)
	os.WriteFile(f2, []byte(`
packages:
  - name: tool
    source: flatpak
`), 0o644)
	_, err := LoadDefaultsAndFiles(nil, []string{f1, f2})
	if err == nil {
		t.Fatalf("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "a.yaml") || !strings.Contains(err.Error(), "b.yaml") {
		t.Fatalf("error should mention both files, got: %v", err)
	}
}

func TestLoadDefaultsAndFiles_DuplicateWithDefaults_ErrorMentionsDefaults(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "u.yaml")
	os.WriteFile(f, []byte(`
packages:
  - name: tool
    source: apt
`), 0o644)
	defaults := []byte(`
packages:
  - name: tool
    source: flatpak
`)
	_, err := LoadDefaultsAndFiles(defaults, []string{f})
	if err == nil {
		t.Fatalf("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "defaults") || !strings.Contains(err.Error(), "u.yaml") {
		t.Fatalf("error should mention defaults and user file, got: %v", err)
	}
}

func TestLoadDefaultsAndFiles_OverlaySources(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "user.yaml")
	os.WriteFile(f, []byte(`
sources:
  - name: apt
    install:
      command: "custom i {package_list}"
      require_root: true
    search:
      command: "s {query}"
`), 0o644)

    defaults := []byte(`
sources:
  - type: package_manager
    name: apt
    install:
      command: "apt install {package_list}"
    remove:
      command: "apt remove {package_list}"
`)

    cfg, err := LoadDefaultsAndFiles(defaults, []string{f})
    if err != nil {
        t.Fatalf("unexpected err: %v", err)
    }
    if len(cfg.Sources) != 1 {
        t.Fatalf("want 1 source, got %d", len(cfg.Sources))
    }
    s := cfg.Sources[0]
    if s.Name != "apt" {
        t.Fatalf("unexpected name: %s", s.Name)
    }
    if s.Type != "package_manager" {
        t.Fatalf("type not preserved from defaults: %s", s.Type)
    }
    if s.Install.Command != "custom i {package_list}" {
        t.Fatalf("install command not overridden: %s", s.Install.Command)
    }
    if s.Install.RequireRoot == nil || *s.Install.RequireRoot != true {
        t.Fatalf("require_root not set true")
    }
    if s.Remove.Command != "apt remove {package_list}" {
        t.Fatalf("remove lost from defaults: %s", s.Remove.Command)
    }
}

// packages overlay is not supported anymore; duplicates must error, so no test here

func TestLoadDefaultsAndFiles_DefaultsOnly(t *testing.T) {
    defaults := []byte(`
sources:
  - name: snap
    type: package_manager
    install:
      command: "snap i {package_list}"
`)
    cfg, err := LoadDefaultsAndFiles(defaults, nil)
    if err != nil {
        t.Fatalf("unexpected err: %v", err)
    }
    if len(cfg.Sources) != 1 || cfg.Sources[0].Name != "snap" {
        t.Fatalf("defaults not loaded correctly")
    }
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
	if err == nil {
		t.Fatalf("expected duplicate error")
	}
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
	if err == nil {
		t.Fatalf("expected duplicate error")
	}
}
