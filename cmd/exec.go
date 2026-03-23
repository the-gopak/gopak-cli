package cmd

import (
	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/spf13/cobra"
)

func init() {
	var noCache bool
	cmd := &cobra.Command{
		Use:   "exec -- <package> [args...]",
		Short: "Update the package if needed, then run it",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			return m.Exec(args[0], args[1:], noCache, cfg.ParsedExecCacheTTL())
		},
	}
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "bypass update-check cache and force a fresh check")
	rootCmd.AddCommand(cmd)
}
