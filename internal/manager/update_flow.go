package manager

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/viktorprogger/universal-linux-installer/internal/config"
	"github.com/viktorprogger/universal-linux-installer/internal/executil"
)

type pkgKey struct {
	Source string
	Name   string
	Kind   string 
}

type pkgStatus struct {
	Installed string
	Available string
}

func (m *Manager) UpdateAll(ctx context.Context, r UpdateReporter, runner Runner) error {
	groups := groupTracked(m.cfg)
	r.OnInit(groups)

	status := make(map[pkgKey]pkgStatus)
	var mu sync.Mutex
	var wg sync.WaitGroup
	allKeys := make([]pkgKey, 0)
	for src, names := range groups {
		for _, n := range names {
			k := pkgKey{Source: src, Name: n, Kind: kindOf(src)}
			allKeys = append(allKeys, k)
			wg.Add(1)
			go func(k pkgKey) {
				defer wg.Done()
				inst := m.getInstalled(k)
				mu.Lock()
				s := status[k]
				s.Installed = inst
				status[k] = s
				mu.Unlock()
				r.OnInstalled(PackageKey{Source: k.Source, Name: k.Name, Kind: k.Kind}, inst)
			}(k)
		}
	}
	wg.Wait()
	r.OnPhaseDone("installed")

	wg = sync.WaitGroup{}
	for _, k := range allKeys {
		wg.Add(1)
		go func(k pkgKey) {
			defer wg.Done()
			avail := m.getAvailable(k)
			mu.Lock()
			s := status[k]
			s.Available = avail
			status[k] = s
			mu.Unlock()
			r.OnAvailable(PackageKey{Source: k.Source, Name: k.Name, Kind: k.Kind}, avail)
		}(k)
	}
	wg.Wait()
	r.OnPhaseDone("available")

	if !r.ConfirmProceed() {
		r.OnDone()
		return nil
	}

	r.OnUpdateStart()
	var runWg sync.WaitGroup
	for src, names := range m.groupSourcesOnly() {
		src := src
		names := append([]string{}, names...)
		s := m.sourceByName(src)
		if len(names) == 0 || s.Name == "" { continue }
		if s.Update.Command == "" { return fmt.Errorf("missing update command for source: %s", src) }
		runWg.Add(1)
		go func() {
			defer runWg.Done()
			cmd := strings.ReplaceAll(s.Update.Command, "{package_list}", strings.Join(names, " "))
			err := runner.Run(src, "update-group", cmd, s.Update.RequireRoot)
			for _, n := range names {
				ok := err == nil
				msg := ""
				if err != nil { msg = err.Error() }
				r.OnPackageUpdated(PackageKey{Source: src, Name: n, Kind: kindOf(src)}, ok, msg)
			}
		}()
	}
	for _, cp := range m.cfg.CustomPackages {
		cp := cp
		runWg.Add(1)
		go func() {
			defer runWg.Done()
			_ = m.updateCustomWithRunner(cp, runner, r)
		}()
	}
	runWg.Wait()
	r.OnDone()
	return nil
}

func (m *Manager) groupSourcesOnly() map[string][]string {
	res := map[string][]string{}
	for _, p := range m.cfg.Packages {
		res[p.Source] = append(res[p.Source], p.Name)
	}
	return res
}

func kindOf(group string) string {
	if group == "custom" { return "custom" }
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

func (m *Manager) getInstalled(k pkgKey) string {
	if k.Kind == "custom" {
		cp := m.customByName(k.Name)
		if cp.GetInstalledVersion.Command == "" { return "" }
		res := executil.RunShell(cp.GetInstalledVersion.Command)
		if res.Code != 0 { return "" }
		return strings.TrimSpace(res.Stdout)
	}
	src := m.sourceByName(k.Source)
	switch src.Name {
	case "apt":
		cmd := fmt.Sprintf("dpkg-query -W -f='${Version}\n' %s 2>/dev/null || true", k.Name)
		res := executil.RunShell(cmd)
		return strings.TrimSpace(res.Stdout)
	default:
		return ""
	}
}

func (m *Manager) getAvailable(k pkgKey) string {
	if k.Kind == "custom" {
		cp := m.customByName(k.Name)
		if cp.GetLatestVersion.Command == "" { return "" }
		res := executil.RunShell(cp.GetLatestVersion.Command)
		if res.Code != 0 { return "" }
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

func (m *Manager) updateCustomWithRunner(cp config.CustomPackage, runner Runner, r UpdateReporter) error {
	need := false
	latest := ""
	installed := ""
	if cp.GetLatestVersion.Command != "" {
		res := executil.RunShell(cp.GetLatestVersion.Command)
		if res.Code != 0 { return fmt.Errorf("command failed for %s [get_latest_version]: exit %d\n%s", cp.Name, res.Code, res.Stderr) }
		latest = strings.TrimSpace(res.Stdout)
	}
	if cp.GetInstalledVersion.Command != "" {
		res := executil.RunShell(cp.GetInstalledVersion.Command)
		if res.Code != 0 { return fmt.Errorf("command failed for %s [get_installed_version]: exit %d\n%s", cp.Name, res.Code, res.Stderr) }
		installed = strings.TrimSpace(res.Stdout)
	}
	if cp.CompareVersions.Command != "" {
		script := fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.CompareVersions.Command)
		res := executil.RunShell(script)
		if res.Code != 0 { return fmt.Errorf("command failed for %s [compare_versions]: exit %d\n%s", cp.Name, res.Code, res.Stderr) }
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
			if err := runner.Run(cp.Name, "download", dl, cp.Download.RequireRoot); err != nil { return err }
		}
		if cp.Remove.Command != "" {
			if err := runner.Run(cp.Name, "remove-before-install", cp.Remove.Command, cp.Remove.RequireRoot); err != nil { return err }
		}
		if cp.Install.Command == "" { return fmt.Errorf("missing install script for custom package: %s", cp.Name) }
		inst := fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.Install.Command)
		if err := runner.Run(cp.Name, "install", inst, cp.Install.RequireRoot); err != nil { return err }
		r.OnPackageUpdated(PackageKey{Source: "custom", Name: cp.Name, Kind: "custom"}, true, "")
		return nil
	}
	r.OnPackageUpdated(PackageKey{Source: "custom", Name: cp.Name, Kind: "custom"}, true, "")
	return nil
}
