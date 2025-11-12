package console

import (
    "fmt"
    "io"
    "testing"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/viktorprogger/universal-linux-installer/internal/manager"
)

func TestBubbleReporter_ImplementsUpdateReporter(t *testing.T) {
	var _ manager.UpdateReporter = NewConsoleReporter()
}

func TestBubbleReporter_ConfirmProceed_NoUpdates(t *testing.T) {
    r := NewConsoleReporterWithOptions(
        tea.WithOutput(io.Discard),
        tea.WithoutRenderer(),
    )
    r.OnInit(map[string][]string{"custom": {"pkg"}})
    r.OnInstalled(manager.PackageKey{Source: "custom", Name: "pkg", Kind: "custom"}, "1.0.0")
    r.OnAvailable(manager.PackageKey{Source: "custom", Name: "pkg", Kind: "custom"}, "1.0.0")
    if r.ConfirmProceed() {
        t.Fatalf("expected false when no pending updates")
    }
    r.OnDone()
}

func TestBubbleReporter_ConfirmProceed_Y(t *testing.T) {
    pr, pw := io.Pipe()
    r := NewConsoleReporterWithOptions(
        tea.WithInput(pr),
        tea.WithOutput(io.Discard),
        tea.WithoutRenderer(),
        tea.WithoutSignalHandler(),
    )
    r.OnInit(map[string][]string{"custom": {"pkg"}})
    r.OnInstalled(manager.PackageKey{Source: "custom", Name: "pkg", Kind: "custom"}, "")
    r.OnAvailable(manager.PackageKey{Source: "custom", Name: "pkg", Kind: "custom"}, "1.0.0")
    go func() {
        time.Sleep(50 * time.Millisecond)
        fmt.Fprint(pw, "y\n")
        pw.Close()
    }()
    if !r.ConfirmProceed() {
        t.Fatalf("expected true on 'y' input")
    }
    r.OnDone()
    pr.Close()
}

func TestBubbleReporter_ConfirmProceed_N(t *testing.T) {
    pr, pw := io.Pipe()
    r := NewConsoleReporterWithOptions(
        tea.WithInput(pr),
        tea.WithOutput(io.Discard),
        tea.WithoutRenderer(),
        tea.WithoutSignalHandler(),
    )
    r.OnInit(map[string][]string{"custom": {"pkg"}})
    r.OnInstalled(manager.PackageKey{Source: "custom", Name: "pkg", Kind: "custom"}, "")
    r.OnAvailable(manager.PackageKey{Source: "custom", Name: "pkg", Kind: "custom"}, "1.0.0")
    go func() {
        time.Sleep(50 * time.Millisecond)
        fmt.Fprint(pw, "n\n")
        pw.Close()
    }()
    if r.ConfirmProceed() {
        t.Fatalf("expected false on 'n' input")
    }
    r.OnDone()
    pr.Close()
}
