package cmd

import (
	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/gopak/gopak-cli/internal/ui/console"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			ui := console.NewConsoleUI(m)
			return ui.RunRemoveImperative(args[0])
		},
	}
	rootCmd.AddCommand(cmd)
}
