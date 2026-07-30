package executil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/the-gopak/gopak-cli/internal/config"
)

type Result struct {
	Stdout string
	Stderr string
	Code   int
}

func RunShell(c config.Command) Result {
	final := c.Command
	if c.RequireRoot && runtime.GOOS != "windows" && os.Geteuid() != 0 {
		esc := strings.ReplaceAll(c.Command, "'", "'\"'\"'")
		final = fmt.Sprintf("sudo bash -ceu '%s'", esc)
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", final)
	} else {
		cmd = exec.Command("bash", "-ceu", final)
	}
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	code := 0
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
		} else {
			code = 1
		}
	}
	return Result{Stdout: out.String(), Stderr: errb.String(), Code: code}
}
