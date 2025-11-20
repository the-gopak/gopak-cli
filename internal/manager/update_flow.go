package manager

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/executil"
)

func kindOf(group string) string {
	if group == "custom" {
		return "custom"
	}
	return "source"
}

func groupTracked(cfg config.Config) map[string][]string {
	res := map[string][]string{}
	for _, p := range cfg.Packages {
		res[p.Source] = append(res[p.Source], p.Name)
	}

	if len(cfg.CustomPackages) > 0 {
		for _, c := range cfg.CustomPackages {
			res["custom"] = append(res["custom"], c.Name)
		}
	}

	for k := range res {
		sort.Strings(res[k])
	}

	return res
}

func (m *Manager) getVersionInstalled(k PackageKey) string {
	if k.Kind == "custom" {
		cp := m.customByName(k.Name)
		if cp.GetInstalledVersion.Command == "" {
			return ""
		}
		res := executil.RunShell(cp.GetInstalledVersion)
		if res.Code != 0 {
			return ""
		}
		return strings.TrimSpace(res.Stdout)
	}
	src := m.sourceByName(k.Source)
	if src.Name == "" {
		return ""
	}
	if src.GetInstalledVersion.Command != "" {
		cmd := strings.ReplaceAll(src.GetInstalledVersion.Command, "{package}", k.Name)
		res := executil.RunShell(config.Command{Command: cmd, RequireRoot: src.GetInstalledVersion.RequireRoot})
		if res.Code != 0 {
			return ""
		}
		return strings.TrimSpace(res.Stdout)
	}
	switch src.Name {
	case "apt":
		cmd := fmt.Sprintf("dpkg-query -W -f='${Version}\n' %s 2>/dev/null || true", k.Name)
		res := executil.RunShell(config.Command{Command: cmd})
		return strings.TrimSpace(res.Stdout)
	default:
		return ""
	}
}

func (m *Manager) getVersionAvailable(k PackageKey) string {
	if k.Kind == "custom" {
		cp := m.customByName(k.Name)
		if cp.GetLatestVersion.Command == "" {
			return ""
		}
		res := executil.RunShell(cp.GetLatestVersion)
		if res.Code != 0 {
			return ""
		}
		return strings.TrimSpace(res.Stdout)
	}

	src := m.sourceByName(k.Source)
	m.ensurePreUpdate(src)
	if src.Name == "" {
		return ""
	}
	if src.GetLatestVersion.Command != "" {
		cmd := strings.ReplaceAll(src.GetLatestVersion.Command, "{package}", k.Name)
		res := executil.RunShell(config.Command{Command: cmd, RequireRoot: src.GetLatestVersion.RequireRoot})
		if res.Code != 0 {
			return ""
		}
		return strings.TrimSpace(res.Stdout)
	}
	switch src.Name {
	case "apt":
		cmd := fmt.Sprintf("apt-cache policy %s | awk '/Candidate:/ {print $2}'", k.Name)
		res := executil.RunShell(config.Command{Command: cmd})
		return strings.TrimSpace(res.Stdout)
	default:
		return ""
	}
}

func (m *Manager) updateCustomWithRunner(cp config.CustomPackage, runner Runner) error {
	need := false
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

	if installed == "" && cp.Install.Command != "" {
		need = true
	} else if latest != "" {
		need = cmpVersion(latest, installed) > 0
	} else {
		need = false
	}

	if need {
		if cp.Remove.Command != "" {
			if err := runner.Run(cp.Name, "remove-before-install", cp.Remove); err != nil {
				return err
			}
		}
		if cp.Install.Command == "" {
			return fmt.Errorf("missing install script for custom package: %s", cp.Name)
		}
		instCmd := config.Command{Command: fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.Install.Command), RequireRoot: cp.Install.RequireRoot}
		if err := runner.Run(cp.Name, "install", instCmd); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (m *Manager) GetVersionInstalled(k PackageKey) string {
	return m.getVersionInstalled(k)
}

func (m *Manager) GetVersionAvailable(k PackageKey) string {
	return m.getVersionAvailable(k)
}

func (m *Manager) Tracked() map[string][]string {
	return groupTracked(m.cfg)
}

func (m *Manager) UpdateSelected(keys []PackageKey, runner Runner, onUpdate func(PackageKey, bool, string)) error {
	bySrcInstall := map[string][]string{}
	bySrcUpdate := map[string][]string{}
	customSet := map[string]struct{}{}
	// classify per package
	for _, k := range keys {
		if k.Kind == "custom" {
			customSet[k.Name] = struct{}{}
			continue
		}
		installed := m.getVersionInstalled(k)
		if installed == "" {
			bySrcInstall[k.Source] = append(bySrcInstall[k.Source], k.Name)
		} else {
			bySrcUpdate[k.Source] = append(bySrcUpdate[k.Source], k.Name)
		}
	}

	var wg sync.WaitGroup
	// run installs per source
	for src, names := range bySrcInstall {
		src, names := src, append([]string{}, names...)
		s := m.sourceByName(src)
		if len(names) == 0 || s.Name == "" {
			continue
		}
		if s.Install.Command == "" {
			return fmt.Errorf("missing install command for source: %s", src)
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
				if onUpdate != nil {
					onUpdate(PackageKey{Source: src, Name: n, Kind: kindOf(src)}, ok, msg)
				}
			}
		}()
	}
	// run updates per source
	for src, names := range bySrcUpdate {
		src, names := src, append([]string{}, names...)
		s := m.sourceByName(src)
		if len(names) == 0 || s.Name == "" {
			continue
		}
		if s.Update.Command == "" {
			return fmt.Errorf("missing update command for source: %s", src)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmdStr := strings.ReplaceAll(s.Update.Command, "{package_list}", strings.Join(names, " "))
			cmd := config.Command{Command: cmdStr, RequireRoot: s.Update.RequireRoot}
			err := runner.Run(src, "update-group", cmd)
			for _, n := range names {
				ok := err == nil
				msg := "updated"
				if err != nil {
					msg = err.Error()
				}
				if onUpdate != nil {
					onUpdate(PackageKey{Source: src, Name: n, Kind: kindOf(src)}, ok, msg)
				}
			}
		}()
	}
	// custom
	for name := range customSet {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.updateCustomWithRunner(m.customByName(name), runner); err != nil {
				if onUpdate != nil {
					onUpdate(PackageKey{Source: "custom", Name: name, Kind: "custom"}, false, err.Error())
				}
			} else {
				if onUpdate != nil {
					onUpdate(PackageKey{Source: "custom", Name: name, Kind: "custom"}, true, "")
				}
			}
		}()
	}
	wg.Wait()
	return nil
}
