package cmd

import (
	"fmt"
	"sort"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/gopak/gopak-cli/internal/ui/console"
	"github.com/spf13/cobra"
)

func init() {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Install one package or select from uninstalled",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			m := manager.New(cfg)
			if dryRun {
				if len(args) == 1 {
					keys, err := m.ResolveKeys(args[0])
					if err != nil {
						return err
					}
					for _, k := range keys {
						ins := m.GetVersionInstalled(k)
						av := m.GetVersionAvailableDryRun(k)
						if ins != "" {
							fmt.Printf("skip (already installed): %s/%s %s\n", k.Source, k.Name, manager.NormalizeVersion(ins))
							continue
						}
						if av != "" {
							fmt.Printf("install: %s/%s -> %s\n", k.Source, k.Name, manager.NormalizeVersion(av))
						} else {
							fmt.Printf("install: %s/%s\n", k.Source, k.Name)
						}
					}
					return nil
				}
				groups := m.Tracked()
				keys := make([]manager.PackageKey, 0)
				for grp, names := range groups {
					for _, n := range names {
						k := manager.PackageKey{Source: grp, Name: n, Kind: manager.KindOf(grp)}
						if !m.HasCommand(k, manager.OpInstall) {
							continue
						}
						if m.GetVersionInstalled(k) != "" {
							continue
						}
						keys = append(keys, k)
					}
				}
				sort.Slice(keys, func(i, j int) bool {
					if keys[i].Source == keys[j].Source {
						return keys[i].Name < keys[j].Name
					}
					return keys[i].Source < keys[j].Source
				})
				if len(keys) == 0 {
					fmt.Println("Nothing to install")
					return nil
				}
				for _, k := range keys {
					av := m.GetVersionAvailableDryRun(k)
					if av != "" {
						fmt.Printf("install: %s/%s -> %s\n", k.Source, k.Name, manager.NormalizeVersion(av))
					} else {
						fmt.Printf("install: %s/%s\n", k.Source, k.Name)
					}
				}
				return nil
			}
			if len(args) == 1 {
				return m.Install(args[0])
			}
			ui := console.NewConsoleUI(m)

			return ui.Install()
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned changes without executing")
	rootCmd.AddCommand(cmd)
}
