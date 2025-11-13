package cmd

import (
	"github.com/spf13/cobra"
	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/gopak/gopak-cli/internal/ui/console"
)

func init() {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update one package or all",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			if len(args) == 1 {
				return m.UpdateOne(args[0])
			}
			ui := console.NewConsoleUI(m)

			return ui.Update()
		},
	}
	rootCmd.AddCommand(cmd)
}
