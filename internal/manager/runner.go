package manager

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Runner interface {
	Run(name, step, script string, require *bool) error
	Close() error
}

type SudoRunner struct {
	keepOnce sync.Once
	stopCh   chan struct{}
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

func (r *SudoRunner) Run(name, step, script string, require *bool) error {
	needRoot := false
	if require != nil && *require { needRoot = true }
	final := script
	if needRoot {
		if os.Geteuid() != 0 {
			vcmd := exec.Command("sudo", "-v")
			vcmd.Stdin = os.Stdin
			if err := vcmd.Run(); err != nil {
				return fmt.Errorf("sudo auth failed for %s [%s]: %v", name, step, err)
			}
			r.ensureKeepAliveStarted()
			esc := strings.ReplaceAll(script, "'", "'\"'\"'")
			final = fmt.Sprintf("sudo bash -ceu '%s'", esc)
		}
	}
	cmd := exec.Command("bash", "-ceu", final)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
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
