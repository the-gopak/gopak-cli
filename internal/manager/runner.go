package manager

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gopak/gopak-cli/internal/config"
)

type Runner interface {
	Run(name, step string, cmd config.Command) error
	Close() error
}

type SudoRunner struct {
	keepOnce sync.Once
	stopCh   chan struct{}
	mu       sync.Mutex
	authed   bool
}

func NewSudoRunner() *SudoRunner { return &SudoRunner{} }

func (r *SudoRunner) ensureKeepAliveStarted() {
	r.keepOnce.Do(func() {
		r.stopCh = make(chan struct{})
		go func() {
			t := time.NewTicker(60 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-r.stopCh:
					return
				case <-t.C:
					_ = exec.Command("sudo", "-n", "-v").Run()
				}
			}
		}()
	})
}

func (r *SudoRunner) ensureRootAccess() bool {
	if os.Geteuid() == 0 {
		return true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.authed {
		return true
	}
	vcmd := exec.Command("sudo", "-v")
	vcmd.Stdin = os.Stdin
	vcmd.Stdout = os.Stdout
	vcmd.Stderr = os.Stderr
	if err := vcmd.Run(); err != nil {
		return false
	}
	r.authed = true
	r.ensureKeepAliveStarted()

	return true
}

func (r *SudoRunner) Run(name, step string, cmd config.Command) error {
	final := cmd.Command
	if cmd.RequireRoot {
		if !r.ensureRootAccess() {
			return fmt.Errorf("sudo auth not granted for %s [%s]", name, step)
		}

		esc := strings.ReplaceAll(final, "'", "'\"'\"'")
		final = fmt.Sprintf("sudo -n bash -ceu '%s'", esc)
	}
	bcmd := exec.Command("bash", "-ceu", final)
	bcmd.Stdout = os.Stdout
	bcmd.Stderr = os.Stderr
	if err := bcmd.Run(); err != nil {
		return fmt.Errorf("command failed for %s [%s]", name, step)
	}
	return nil
}

func (r *SudoRunner) Close() error {
	if r.stopCh != nil {
		close(r.stopCh)
	}
	return nil
}
