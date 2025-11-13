package cmd

import (
	"github.com/spf13/cobra"
	"github.com/viktorprogger/universal-linux-installer/internal/config"
	"github.com/viktorprogger/universal-linux-installer/internal/manager"
	"github.com/viktorprogger/universal-linux-installer/internal/ui/console"
)

func init() {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search across sources",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			ui := console.NewConsoleUI(m)
			return ui.RunSearchImperative(args[0])
		},
	}
	rootCmd.AddCommand(cmd)
}
