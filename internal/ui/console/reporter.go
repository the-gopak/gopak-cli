package console

import (
	"github.com/gopak/gopak-cli/internal/manager"
)

// ConsoleUI renders and controls the interactive console update flow.
// It queries the manager for tracked packages and their versions, presents a summary,
// asks the user to proceed, and applies updates while streaming results.
type ConsoleUI struct{ m *manager.Manager }

// NewConsoleUI constructs a ConsoleUI bound to the provided manager dependency.
// Use this to run interactive update flows and future console actions.
func NewConsoleUI(m *manager.Manager) *ConsoleUI { return &ConsoleUI{m: m} }

func kindOf(group string) string {
	if group == "custom" {
		return "custom"
	}
	return "source"
}
