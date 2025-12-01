package manager

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/executil"
)

type Operation string

const (
	OpInstall Operation = "install"
	OpUpdate  Operation = "update"
)

func KindOf(group string) string {
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
	return ""
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
	return ""
}

func (m *Manager) executeCustomWithRunner(cp config.CustomPackage, op Operation, runner Runner) error {
	var cmd config.Command
	switch op {
	case OpInstall:
		cmd = cp.Install
	case OpUpdate:
		cmd = cp.Update
	}
	if cmd.Command == "" {
		if op == OpInstall {
			return fmt.Errorf("missing install script for custom package: %s", cp.Name)
		}
		return nil
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

	need := false
	switch op {
	case OpInstall:
		need = installed == ""
	case OpUpdate:
		if installed == "" {
			return nil
		}
		if latest != "" {
			need = cmpVersion(latest, installed) > 0
		}
	}

	if !need {
		return nil
	}

	if op == OpInstall && cp.Remove.Command != "" {
		if err := runner.Run(cp.Name, "remove-before-install", cp.Remove); err != nil {
			return err
		}
	}

	execCmd := config.Command{
		Command:     fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cmd.Command),
		RequireRoot: cmd.RequireRoot,
	}
	return runner.Run(cp.Name, string(op), execCmd)
}

func (m *Manager) GetVersionInstalled(k PackageKey) string {
	return m.getVersionInstalled(k)
}

func (m *Manager) GetVersionAvailable(k PackageKey) string {
	return m.getVersionAvailable(k)
}

func (m *Manager) HasCommand(k PackageKey, op Operation) bool {
	if k.Kind == "custom" {
		cp := m.customByName(k.Name)
		switch op {
		case OpInstall:
			return cp.Install.Command != ""
		case OpUpdate:
			return cp.Update.Command != ""
		}
		return false
	}
	src := m.sourceByName(k.Source)
	if src.Name == "" {
		return false
	}
	switch op {
	case OpInstall:
		return src.Install.Command != ""
	case OpUpdate:
		return src.Update.Command != ""
	}
	return false
}

func (m *Manager) Tracked() map[string][]string {
	return groupTracked(m.cfg)
}

func (m *Manager) ExecuteSelected(keys []PackageKey, op Operation, runner Runner, onDone func(PackageKey, bool, string)) error {
	bySrc := map[string][]string{}
	customSet := map[string]struct{}{}
	for _, k := range keys {
		if k.Kind == "custom" {
			customSet[k.Name] = struct{}{}
			continue
		}
		bySrc[k.Source] = append(bySrc[k.Source], k.Name)
	}

	var wg sync.WaitGroup
	for src, names := range bySrc {
		src, names := src, append([]string{}, names...)
		s := m.sourceByName(src)
		if len(names) == 0 || s.Name == "" {
			continue
		}
		var srcCmd config.Command
		switch op {
		case OpInstall:
			srcCmd = s.Install
		case OpUpdate:
			srcCmd = s.Update
		}
		if srcCmd.Command == "" {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmdStr := strings.ReplaceAll(srcCmd.Command, "{package_list}", strings.Join(names, " "))
			cmd := config.Command{Command: cmdStr, RequireRoot: srcCmd.RequireRoot}
			err := runner.Run(src, string(op)+"-group", cmd)
			for _, n := range names {
				ok := err == nil
				msg := "updated"
				if op == OpInstall {
					msg = "installed"
				}
				if err != nil {
					msg = err.Error()
				}
				if onDone != nil {
					onDone(PackageKey{Source: src, Name: n, Kind: KindOf(src)}, ok, msg)
				}
			}
		}()
	}
	for name := range customSet {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := "updated"
			if op == OpInstall {
				msg = "installed"
			}
			if err := m.executeCustomWithRunner(m.customByName(name), op, runner); err != nil {
				if onDone != nil {
					onDone(PackageKey{Source: "custom", Name: name, Kind: "custom"}, false, err.Error())
				}
			} else {
				if onDone != nil {
					onDone(PackageKey{Source: "custom", Name: name, Kind: "custom"}, true, msg)
				}
			}
		}()
	}
	wg.Wait()
	return nil
}

func (m *Manager) UpdateSelected(keys []PackageKey, runner Runner, onUpdate func(PackageKey, bool, string)) error {
	return m.ExecuteSelected(keys, OpUpdate, runner, onUpdate)
}

func (m *Manager) InstallSelected(keys []PackageKey, runner Runner, onInstall func(PackageKey, bool, string)) error {
	return m.ExecuteSelected(keys, OpInstall, runner, onInstall)
}
