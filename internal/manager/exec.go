package manager

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gopak/gopak-cli/internal/logging"
)

func (m *Manager) executableForPackage(name string) string {
	if cp := m.customByName(name); cp.Name != "" && cp.Executable != "" {
		return cp.Executable
	}
	if gp := m.githubByName(name); gp.Name != "" && gp.Executable != "" {
		return gp.Executable
	}
	if p := m.pkgByName(name); p.Name != "" && p.Executable != "" {
		return p.Executable
	}
	return name
}

func (m *Manager) Exec(packageName string, extraArgs []string, noCache bool, ttl time.Duration) error {
	if !m.isCustom(packageName) && !m.isGithubRelease(packageName) && m.pkgByName(packageName).Name == "" {
		return fmt.Errorf("Package %q is not known by Gopak", packageName)
	}

	lock, err := acquireExecLock(packageName)
	if err != nil {
		logging.Debug(fmt.Sprintf("exec: could not acquire lock for %s: %v", packageName, err))
	}
	if lock != nil {
		defer lock.Close()
	}

	doCheck := noCache
	if !noCache {
		cache := loadExecCache()
		doCheck = !cache.IsFresh(packageName, ttl)
	}
	if doCheck {
		if err := m.UpdateOne(packageName); err != nil {
			logging.Debug(fmt.Sprintf("exec: update failed for %s: %v", packageName, err))
		}
		cache := loadExecCache()
		cache.Touch(packageName)
		saveExecCache(cache)
	}

	if lock != nil {
		_ = lock.Close()
		lock = nil
	}

	binary := m.executableForPackage(packageName)
	child := exec.Command(binary, extraArgs...)
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	err = child.Run()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			os.Exit(e.ExitCode())
		}
		return err
	}
	return nil
}
