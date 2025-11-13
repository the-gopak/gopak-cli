package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var logfile *os.File
var verbose bool

func Init() {
	dir, _ := os.UserConfigDir()
	p := filepath.Join(dir, "unilin", "logs")
	_ = os.MkdirAll(p, 0o755)
	f, _ := os.OpenFile(filepath.Join(p, "unilin.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	logfile = f
	log.SetOutput(f)
}

func Close() {
	if logfile != nil {
		_ = logfile.Close()
	}
}

func color(code, s string) string { return "\x1b[" + code + "m" + s + "\x1b[0m" }

func Info(msg string) {
	fmt.Println(msg)
	log.Println(msg)
}

func Success(msg string) {
	fmt.Println(color("32", msg))
	log.Println(msg)
}

func Error(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, color("31", msg))
	log.Println(msg)
}

func Gray(msg string) {
	fmt.Println(color("90", msg))
	log.Println(msg)
}

// SetVerbose toggles verbose output to stdout.
func SetVerbose(v bool) { verbose = v }

// Debug prints only when verbose mode is enabled.
func Debug(msg string) {
	if !verbose {
		return
	}
	fmt.Println(color("90", msg))
	log.Println("[DEBUG] " + msg)
}
