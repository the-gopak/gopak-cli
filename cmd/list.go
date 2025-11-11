package cmd

import (
	"github.com/spf13/cobra"
	"github.com/viktorprogger/universal-linux-installer/internal/config"
	"github.com/viktorprogger/universal-linux-installer/internal/manager"
)

func init() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)

			return m.List()
		},
	}
	rootCmd.AddCommand(cmd)
}
