package cmd

import (
	"github.com/spf13/cobra"
	"github.com/viktorprogger/universal-linux-installer/internal/config"
	"github.com/viktorprogger/universal-linux-installer/internal/manager"
)

func init() {
	cmd := &cobra.Command{
		Use:   "install <name>",
		Short: "Install a package or custom package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			return m.Install(args[0])
		},
	}
	rootCmd.AddCommand(cmd)
}
