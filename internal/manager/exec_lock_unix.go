//go:build !windows

package manager

import (
	"os"
	"syscall"

	"github.com/gopak/gopak-cli/internal/logging"
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
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		logging.Debug("exec: flock failed: " + err.Error())
		return nil, err
	}
	return f, nil
}
