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
		Use:   "update [name]",
		Short: "Update one package or all",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			ui := console.NewConsoleUI(m)
			return ui.Update(name, dryRun, yes)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned changes without executing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "assume yes and update all without prompting")
	rootCmd.AddCommand(cmd)
}
