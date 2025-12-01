package cmd

import (
	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/gopak/gopak-cli/internal/ui/console"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Install one package or select from uninstalled",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			if len(args) == 1 {
				return m.Install(args[0])
			}
			ui := console.NewConsoleUI(m)

			return ui.Install()
		},
	}
	rootCmd.AddCommand(cmd)
}
