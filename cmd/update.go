package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/viktorprogger/universal-linux-installer/internal/config"
	"github.com/viktorprogger/universal-linux-installer/internal/manager"
	"github.com/viktorprogger/universal-linux-installer/internal/ui/console"
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
			reporter := console.NewConsoleReporter()
			runner := manager.NewSudoRunner()
			defer runner.Close()
			return m.UpdateAll(context.Background(), reporter, runner)
		},
	}
	rootCmd.AddCommand(cmd)
}
