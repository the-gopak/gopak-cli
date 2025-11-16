package manager

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/executil"
)

func (m *Manager) groupSourcesOnly() map[string][]string {
	res := map[string][]string{}
	for _, p := range m.cfg.Packages {
		res[p.Source] = append(res[p.Source], p.Name)
	}
	return res
}

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
		res := executil.RunShell(cp.GetInstalledVersion.Command)
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
		res := executil.RunShell(cmd)
		if res.Code != 0 {
			return ""
		}
		return strings.TrimSpace(res.Stdout)
	}
	switch src.Name {
	case "apt":
		cmd := fmt.Sprintf("dpkg-query -W -f='${Version}\\n' %s 2>/dev/null || true", k.Name)
		res := executil.RunShell(cmd)
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
		res := executil.RunShell(cp.GetLatestVersion.Command)
		if res.Code != 0 {
			return ""
		}
		return strings.TrimSpace(res.Stdout)
	}
	src := m.sourceByName(k.Source)
	switch src.Name {
	case "apt":
		cmd := fmt.Sprintf("apt-cache policy %s | awk '/Candidate:/ {print $2}'", k.Name)
		res := executil.RunShell(cmd)
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
		res := executil.RunShell(cp.GetLatestVersion.Command)
		if res.Code != 0 {
			return fmt.Errorf("command failed for %s [get_latest_version]: exit %d\n%s", cp.Name, res.Code, res.Stderr)
		}
		latest = strings.TrimSpace(res.Stdout)
	}
	if cp.GetInstalledVersion.Command != "" {
		res := executil.RunShell(cp.GetInstalledVersion.Command)
		if res.Code != 0 {
			return fmt.Errorf("command failed for %s [get_installed_version]: exit %d\n%s", cp.Name, res.Code, res.Stderr)
		}
		installed = strings.TrimSpace(res.Stdout)
	}
	if cp.CompareVersions.Command != "" {
		script := fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.CompareVersions.Command)
		res := executil.RunShell(script)
		if res.Code != 0 {
			return fmt.Errorf("command failed for %s [compare_versions]: exit %d\n%s", cp.Name, res.Code, res.Stderr)
		}
		out := strings.ToLower(strings.TrimSpace(res.Stdout))
		need = out == "true" || out == "1" || out == "yes"
	} else {
		if installed == "" && cp.Install.Command != "" {
			need = true
		} else if latest != "" {
			need = cmpVersion(latest, installed) > 0
		} else {
			need = false
		}
	}
	if need {
		if cp.Download.Command != "" {
			dl := fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.Download.Command)
			if err := runner.Run(cp.Name, "download", dl, cp.Download.RequireRoot); err != nil {
				return err
			}
		}
		if cp.Remove.Command != "" {
			if err := runner.Run(cp.Name, "remove-before-install", cp.Remove.Command, cp.Remove.RequireRoot); err != nil {
				return err
			}
		}
		if cp.Install.Command == "" {
			return fmt.Errorf("missing install script for custom package: %s", cp.Name)
		}
		inst := fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.Install.Command)
		if err := runner.Run(cp.Name, "install", inst, cp.Install.RequireRoot); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (m *Manager) Tracked() map[string][]string {
	return groupTracked(m.cfg)
}

func (m *Manager) GetVersionsInstalled(keys []PackageKey) map[PackageKey]string {
	out := make(map[PackageKey]string, len(keys))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, k := range keys {
		k := k
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := m.getVersionInstalled(k)
			mu.Lock()
			out[k] = v
			mu.Unlock()
		}()
	}
	wg.Wait()
	return out
}

func (m *Manager) GetVersionsAvailable(keys []PackageKey) map[PackageKey]string {
	out := make(map[PackageKey]string, len(keys))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, k := range keys {
		k := k
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := m.getVersionAvailable(k)
			mu.Lock()
			out[k] = v
			mu.Unlock()
		}()
	}
	wg.Wait()
	return out
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
			cmd := strings.ReplaceAll(s.Install.Command, "{package_list}", strings.Join(names, " "))
			err := runner.Run(src, "install-group", cmd, s.Install.RequireRoot)
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
			cmd := strings.ReplaceAll(s.Update.Command, "{package_list}", strings.Join(names, " "))
			err := runner.Run(src, "update-group", cmd, s.Update.RequireRoot)
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
