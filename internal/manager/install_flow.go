package manager

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/executil"
)

func (m *Manager) installCustomWithRunner(cp config.CustomPackage, runner Runner) error {
	if cp.Install.Command == "" {
		return fmt.Errorf("missing install script for custom package: %s", cp.Name)
	}
	latest := ""
	installed := ""
	if cp.GetLatestVersion.Command != "" {
		res := executil.RunShell(cp.GetLatestVersion)
		if res.Code != 0 {
			return fmt.Errorf("command failed for %s [get_latest_version]: exit %d\n%s", cp.Name, res.Code, res.Stderr)
		}
		latest = strings.TrimSpace(res.Stdout)
	}
	if cp.GetInstalledVersion.Command != "" {
		res := executil.RunShell(cp.GetInstalledVersion)
		if res.Code != 0 {
			return fmt.Errorf("command failed for %s [get_installed_version]: exit %d\n%s", cp.Name, res.Code, res.Stderr)
		}
		installed = strings.TrimSpace(res.Stdout)
	}

	if installed != "" {
		return nil
	}

	if cp.Remove.Command != "" {
		if err := runner.Run(cp.Name, "remove-before-install", cp.Remove); err != nil {
			return err
		}
	}
	instCmd := config.Command{Command: fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.Install.Command), RequireRoot: cp.Install.RequireRoot}
	return runner.Run(cp.Name, "install", instCmd)
}

func (m *Manager) InstallSelected(keys []PackageKey, runner Runner, onInstall func(PackageKey, bool, string)) error {
	bySrcInstall := map[string][]string{}
	customSet := map[string]struct{}{}
	for _, k := range keys {
		if k.Kind == "custom" {
			customSet[k.Name] = struct{}{}
			continue
		}
		bySrcInstall[k.Source] = append(bySrcInstall[k.Source], k.Name)
	}

	var wg sync.WaitGroup
	for src, names := range bySrcInstall {
		src, names := src, append([]string{}, names...)
		s := m.sourceByName(src)
		if len(names) == 0 || s.Name == "" {
			continue
		}
		if s.Install.Command == "" {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmdStr := strings.ReplaceAll(s.Install.Command, "{package_list}", strings.Join(names, " "))
			cmd := config.Command{Command: cmdStr, RequireRoot: s.Install.RequireRoot}
			err := runner.Run(src, "install-group", cmd)
			for _, n := range names {
				ok := err == nil
				msg := "installed"
				if err != nil {
					msg = err.Error()
				}
				if onInstall != nil {
					onInstall(PackageKey{Source: src, Name: n, Kind: kindOf(src)}, ok, msg)
				}
			}
		}()
	}
	for name := range customSet {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.installCustomWithRunner(m.customByName(name), runner); err != nil {
				if onInstall != nil {
					onInstall(PackageKey{Source: "custom", Name: name, Kind: "custom"}, false, err.Error())
				}
			} else {
				if onInstall != nil {
					onInstall(PackageKey{Source: "custom", Name: name, Kind: "custom"}, true, "installed")
				}
			}
		}()
	}
	wg.Wait()
	return nil
}
