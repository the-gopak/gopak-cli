//go:build windows

package manager

import (
	"os"

	"github.com/gopak/gopak-cli/internal/logging"
	"golang.org/x/sys/windows"
)

func acquireExecLock(pkg string) (*os.File, error) {
	dir := execCacheDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(execLockPath(pkg), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	// LockFileEx with LOCKFILE_EXCLUSIVE_LOCK (no LOCKFILE_FAIL_IMMEDIATELY) blocks
	// until the lock is available — same semantics as flock(LOCK_EX) on Unix.
	ol := new(windows.Overlapped)
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, ol); err != nil {
		_ = f.Close()
		logging.Debug("exec: LockFileEx failed: " + err.Error())
		return nil, err
	}
	return f, nil
}
