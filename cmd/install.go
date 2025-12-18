package cmd

import (
	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/gopak/gopak-cli/internal/ui/console"
	"github.com/spf13/cobra"
)

func init() {
	var dryRun bool
	var yes bool
	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Install one package or select from uninstalled",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			ui := console.NewConsoleUI(m)
			return ui.Install(name, dryRun, yes)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned changes without executing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "assume yes and install all without prompting")
	rootCmd.AddCommand(cmd)
}
