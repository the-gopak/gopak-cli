package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type PackageState struct {
	Version       string            `json:"version"`
	InstalledAt   string            `json:"installed_at"`
	FileChecksums map[string]string `json:"file_checksums,omitempty"`
}

type State struct {
	Packages map[string]PackageState `json:"packages"`
}

type Manager struct {
	path  string
	state State
	mu    sync.RWMutex
}

func NewManager(configDir string) (*Manager, error) {
	path := filepath.Join(configDir, "state.json")
	m := &Manager{
		path:  path,
		state: State{Packages: make(map[string]PackageState)},
	}
	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return m, nil
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.state)
}

func (m *Manager) save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0o644)
}

func (m *Manager) GetPackageState(name string) (PackageState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ps, ok := m.state.Packages[name]
	return ps, ok
}

func (m *Manager) SetPackageState(name string, ps PackageState) error {
	m.mu.Lock()
	m.state.Packages[name] = ps
	m.mu.Unlock()
	return m.save()
}

func (m *Manager) RemovePackageState(name string) error {
	m.mu.Lock()
	delete(m.state.Packages, name)
	m.mu.Unlock()
	return m.save()
}

func (m *Manager) VerifyChecksums(name string, files []string) (bool, error) {
	m.mu.RLock()
	ps, ok := m.state.Packages[name]
	m.mu.RUnlock()
	if !ok || len(ps.FileChecksums) == 0 {
		return false, nil
	}

	for _, file := range files {
		expectedSum, exists := ps.FileChecksums[file]
		if !exists {
			continue
		}
		actualSum, err := FileChecksum(file)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if actualSum != expectedSum {
			return false, nil
		}
	}
	return true, nil
}

func FileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
