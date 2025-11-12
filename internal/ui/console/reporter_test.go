package console

import (
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/viktorprogger/universal-linux-installer/internal/manager"
)

func TestBubbleReporter_ImplementsUpdateReporter(t *testing.T) {
	var _ manager.UpdateReporter = NewConsoleReporter()
}

func TestBubbleReporter_ConfirmProceed_Y(t *testing.T) {
	r := NewConsoleReporterWithOptions(
		tea.WithInput(strings.NewReader("y\n")),
		tea.WithOutput(io.Discard),
		tea.WithoutRenderer(),
	)
	if !r.ConfirmProceed() {
		t.Fatalf("expected true on 'y' input")
	}
	r.OnDone()
}

func TestBubbleReporter_ConfirmProceed_N(t *testing.T) {
	r := NewConsoleReporterWithOptions(
		tea.WithInput(strings.NewReader("n\n")),
		tea.WithOutput(io.Discard),
		tea.WithoutRenderer(),
	)
	if r.ConfirmProceed() {
		t.Fatalf("expected false on 'n' input")
	}
	r.OnDone()
}
