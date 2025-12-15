package cmd

import (
	"fmt"
	"sort"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/gopak/gopak-cli/internal/ui/console"
	"github.com/spf13/cobra"
)

func updateNeeded(installed, available string) bool {
	if installed == "" || available == "" {
		return false
	}
	ni := manager.NormalizeVersion(installed)
	na := manager.NormalizeVersion(available)
	if ni == "" || na == "" {
		return installed != available
	}
	return manager.CompareVersions(na, ni) > 0
}

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
			if dryRun {
				if len(args) == 1 {
					k, err := m.KeyForName(args[0])
					if err != nil {
						return err
					}
					ins := m.GetVersionInstalled(k)
					if ins == "" {
						fmt.Printf("not installed: %s/%s\n", k.Source, k.Name)
						return nil
					}
					av := m.GetVersionAvailableDryRun(k)
					if updateNeeded(ins, av) {
						fmt.Printf("update: %s/%s %s -> %s\n", k.Source, k.Name, manager.NormalizeVersion(ins), manager.NormalizeVersion(av))
						return nil
					}
					fmt.Printf("up-to-date: %s/%s %s\n", k.Source, k.Name, manager.NormalizeVersion(ins))
					return nil
				}
				groups := m.Tracked()
				keys := make([]manager.PackageKey, 0)
				for grp, names := range groups {
					for _, n := range names {
						k := manager.PackageKey{Source: grp, Name: n, Kind: manager.KindOf(grp)}
						if !m.HasCommand(k, manager.OpUpdate) {
							continue
						}
						ins := m.GetVersionInstalled(k)
						if ins == "" {
							continue
						}
						av := m.GetVersionAvailableDryRun(k)
						if updateNeeded(ins, av) {
							keys = append(keys, k)
						}
					}
				}
				sort.Slice(keys, func(i, j int) bool {
					if keys[i].Source == keys[j].Source {
						return keys[i].Name < keys[j].Name
					}
					return keys[i].Source < keys[j].Source
				})
				if len(keys) == 0 {
					fmt.Println("Nothing to update")
					return nil
				}
				for _, k := range keys {
					ins := m.GetVersionInstalled(k)
					av := m.GetVersionAvailableDryRun(k)
					fmt.Printf("update: %s/%s %s -> %s\n", k.Source, k.Name, manager.NormalizeVersion(ins), manager.NormalizeVersion(av))
				}
				return nil
			}
			if len(args) == 1 {
				return m.UpdateOne(args[0])
			}
			if yes {
				groups := m.Tracked()
				keys := make([]manager.PackageKey, 0)
				for grp, names := range groups {
					for _, n := range names {
						k := manager.PackageKey{Source: grp, Name: n, Kind: manager.KindOf(grp)}
						if !m.HasCommand(k, manager.OpUpdate) {
							continue
						}
						ins := m.GetVersionInstalled(k)
						if ins == "" {
							continue
						}
						av := m.GetVersionAvailable(k)
						if updateNeeded(ins, av) {
							keys = append(keys, k)
						}
					}
				}
				sort.Slice(keys, func(i, j int) bool {
					if keys[i].Source == keys[j].Source {
						return keys[i].Name < keys[j].Name
					}
					return keys[i].Source < keys[j].Source
				})
				if len(keys) == 0 {
					fmt.Println("Nothing to update")
					return nil
				}
				runner := manager.NewSudoRunner()
				defer runner.Close()
				return m.UpdateSelected(keys, runner, func(k manager.PackageKey, ok bool, msg string) {
					if ok {
						fmt.Println("updated: " + k.Name)
						return
					}
					fmt.Println("failed:  " + k.Name)
					if msg != "" {
						fmt.Println(msg)
					}
				})
			}
			ui := console.NewConsoleUI(m)

			return ui.Update()
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned changes without executing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "assume yes and update all without prompting")
	rootCmd.AddCommand(cmd)
}
