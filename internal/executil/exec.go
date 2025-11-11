package executil

import (
	"bytes"
	"os/exec"
)

type Result struct {
	Stdout string
	Stderr string
	Code   int
}

func RunShell(script string) Result {
	cmd := exec.Command("bash", "-ceu", script)
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
