package manager

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/executil"
	"github.com/gopak/gopak-cli/internal/logging"
)

type Manager struct {
	cfg           config.Config
	customByIdx   map[string]int
	pkgByIdx      map[string]int
	sourceByIdx   map[string]int
	preUpdateOnce sync.Map
}

func hashScript(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func (m *Manager) resetPreUpdateCache() {
	m.preUpdateOnce.Range(func(k, _ any) bool {
		m.preUpdateOnce.Delete(k)
		return true
	})
}

func (m *Manager) ensurePreUpdate(src config.Source) {
	if src.PreUpdate.Command == "" {
		return
	}
	h := hashScript(src.PreUpdate.Command)
	if _, loaded := m.preUpdateOnce.LoadOrStore(h, struct{}{}); loaded {
		return
	}
	logging.Debug(fmt.Sprintf("%s [pre_update]: %s", src.Name, src.PreUpdate.Command))
	res := executil.RunShell(src.PreUpdate)
	if res.Code != 0 {
		logging.Debug(fmt.Sprintf("%s [pre_update failed]: exit=%d", src.Name, res.Code))
	}
}

func cmpVersion(a, b string) int {
	logging.Debug(fmt.Sprintf("cmpVersion input: a=%q b=%q", a, b))
	na := normalizeVersion(a)
	nb := normalizeVersion(b)
	logging.Debug(fmt.Sprintf("cmpVersion normalized: a=%q b=%q", na, nb))
	va := splitNumeric(na)
	vb := splitNumeric(nb)
	logging.Debug(fmt.Sprintf("cmpVersion numeric: a=%v b=%v", va, vb))
	la, lb := len(va), len(vb)
	n := la
	if lb > n {
		n = lb
	}
	for i := 0; i < n; i++ {
		ai := 0
		bi := 0
		if i < la {
			ai = va[i]
		}
		if i < lb {
			bi = vb[i]
		}
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}

func normalizeVersion(s string) string {
	s = strings.TrimSpace(s)
	start := 0
	for start < len(s) && (s[start] < '0' || s[start] > '9') {
		start++
	}
	s = s[start:]
	end := 0
	for end < len(s) {
		c := s[end]
		if (c < '0' || c > '9') && c != '.' {
			break
		}
		end++
	}
	return s[:end]
}

func splitNumeric(s string) []int {
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			out = append(out, 0)
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			n = 0
		}
		out = append(out, n)
	}
	return out
}

func New(cfg config.Config) *Manager {
	m := &Manager{
		cfg:         cfg,
		customByIdx: make(map[string]int, len(cfg.CustomPackages)),
		pkgByIdx:    make(map[string]int, len(cfg.Packages)),
		sourceByIdx: make(map[string]int, len(cfg.Sources)),
	}
	for i, cp := range cfg.CustomPackages {
		m.customByIdx[cp.Name] = i
	}
	for i, p := range cfg.Packages {
		m.pkgByIdx[p.Name] = i
	}
	for i, s := range cfg.Sources {
		m.sourceByIdx[s.Name] = i
	}
	return m
}

func (m *Manager) Install(name string) error {
	plan, err := m.resolve(name)
	if err != nil {
		return err
	}
	logging.Debug(fmt.Sprintf("install plan for %s: %s", name, strings.Join(plan, " -> ")))
	for _, n := range plan {
		if m.isCustom(n) {
			cp := m.customByName(n)
			if cp.Remove.Command != "" {
				if err := m.runCtx(n, "remove-before-install", cp.Remove); err != nil {
					return err
				}
			}
			if cp.Install.Command == "" {
				return fmt.Errorf("missing install script for custom package: %s", n)
			}
			if err := m.runCtx(n, "install", cp.Install); err != nil {
				return err
			}
			logging.Success("installed: " + n)
		} else {
			p := m.pkgByName(n)
			s := m.sourceByName(p.Source)
			cmd := strings.ReplaceAll(s.Install.Command, "{package_list}", n)
			if err := m.runCtx(n, "install", config.Command{Command: cmd, RequireRoot: s.Install.RequireRoot}); err != nil {
				return err
			}
			logging.Success("installed: " + n)
		}
	}
	return nil
}

func (m *Manager) Remove(name string) error {
	if m.isCustom(name) {
		cp := m.customByName(name)
		if cp.Remove.Command == "" {
			return fmt.Errorf("missing remove script for custom package: %s", name)
		}
		return m.runCtx(name, "remove", cp.Remove)
	}
	p := m.pkgByName(name)
	s := m.sourceByName(p.Source)
	cmd := strings.ReplaceAll(s.Remove.Command, "{package_list}", name)
	return m.runCtx(name, "remove", config.Command{Command: cmd, RequireRoot: s.Remove.RequireRoot})
}

func (m *Manager) UpdateOne(name string) error {
	logging.Debug("update one: " + name)
	if m.isCustom(name) {
		return m.updateCustom(m.customByName(name))
	}
	p := m.pkgByName(name)
	s := m.sourceByName(p.Source)
	cmd := strings.ReplaceAll(s.Update.Command, "{package_list}", name)
	if err := m.runCtx(name, "update", config.Command{Command: cmd, RequireRoot: s.Update.RequireRoot}); err != nil {
		return err
	}
	logging.Success("updated: " + name)
	return nil
}

// UpdateAll removed in favor of UpdateAll(ctx, reporter, runner)

func (m *Manager) List() error {
	for _, p := range m.cfg.Packages {
		logging.Info("pkg: " + p.Name)
	}

	for _, cp := range m.cfg.CustomPackages {
		v := ""
		if cp.GetInstalledVersion.Command != "" {
			res := executil.RunShell(cp.GetInstalledVersion)
			v = strings.TrimSpace(res.Stdout)
		}
		if v == "" {
			v = "not installed"
		}
		logging.Info("custom: " + cp.Name + " (" + v + ")")
	}
	return nil
}

func (m *Manager) Search(query string) error {
	for _, s := range m.cfg.Sources {
		if s.Search.Command == "" {
			continue
		}
		cmd := strings.ReplaceAll(s.Search.Command, "{query}", query)
		logging.Debug(fmt.Sprintf("%s [search]: %s", s.Name, cmd))
		res := executil.RunShell(config.Command{Command: cmd, RequireRoot: s.Search.RequireRoot})
		if res.Stdout != "" {
			fmt.Print(res.Stdout)
		}
		if res.Stderr != "" {
			fmt.Print(res.Stderr)
		}
	}
	return nil
}

func (m *Manager) updateCustom(cp config.CustomPackage) error {
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
		logging.Debug(fmt.Sprintf("%s [get_installed_version]: %s", cp.Name, cp.GetInstalledVersion.Command))
		res := executil.RunShell(cp.GetInstalledVersion)
		if res.Code != 0 {
			return fmt.Errorf("command failed for %s [get_installed_version]: exit %d\n%s", cp.Name, res.Code, res.Stderr)
		}
		installed = strings.TrimSpace(res.Stdout)
		logging.Debug(fmt.Sprintf("%s [get_installed_version result]: %s", cp.Name, installed))
	}
	if installed == "" && cp.Install.Command != "" {
		need = true
	} else if latest != "" {
		need = cmpVersion(latest, installed) > 0
	} else {
		need = false
	}
	logging.Debug(fmt.Sprintf("%s versions: latest=%q installed=%q need=%v", cp.Name, latest, installed, need))
	if need {
		if cp.Remove.Command != "" {
			if err := m.runCtx(cp.Name, "remove-before-install", cp.Remove); err != nil {
				return err
			}
		}
		if cp.Install.Command == "" {
			return fmt.Errorf("missing install script for custom package: %s", cp.Name)
		}
		inst := fmt.Sprintf("latest_version=%q installed_version=%q; %s", latest, installed, cp.Install.Command)
		if err := m.runCtx(cp.Name, "install", config.Command{Command: inst, RequireRoot: cp.Install.RequireRoot}); err != nil {
			return err
		}
		logging.Success("updated: " + cp.Name)
		return nil
	}
	logging.Info("up-to-date: " + cp.Name)
	return nil
}

func (m *Manager) resolve(name string) ([]string, error) {
	nodes := map[string][]string{}
	for _, p := range m.cfg.Packages {
		nodes[p.Name] = append([]string{}, p.DependsOn...)
	}
	for _, c := range m.cfg.CustomPackages {
		nodes[c.Name] = append([]string{}, c.DependsOn...)
	}
	if _, ok := nodes[name]; !ok {
		return nil, errors.New("unknown package: " + name)
	}
	ord, ok := topoOrder(nodes)
	if !ok {
		return nil, errors.New("dependency cycle")
	}
	closure := map[string]bool{}
	var visit func(n string)
	visit = func(n string) {
		if closure[n] {
			return
		}
		closure[n] = true
		for _, d := range nodes[n] {
			visit(d)
		}
	}
	visit(name)
	res := []string{}
	for _, n := range ord {
		if closure[n] {
			res = append(res, n)
		}
	}
	return res, nil
}

func (m *Manager) isCustom(name string) bool {
	_, ok := m.customByIdx[name]
	return ok
}

func (m *Manager) customByName(name string) config.CustomPackage {
	if i, ok := m.customByIdx[name]; ok {
		return m.cfg.CustomPackages[i]
	}
	return config.CustomPackage{}
}

func (m *Manager) pkgByName(name string) config.Package {
	if i, ok := m.pkgByIdx[name]; ok {
		return m.cfg.Packages[i]
	}
	return config.Package{}
}

func (m *Manager) sourceByName(name string) config.Source {
	if i, ok := m.sourceByIdx[name]; ok {
		return m.cfg.Sources[i]
	}
	return config.Source{}
}

func (m *Manager) runCtx(name string, step string, command config.Command) error {
	logging.Debug(fmt.Sprintf("%s [%s]: %s", name, step, command.Command))
	res := executil.RunShell(command)
	if res.Stdout != "" {
		fmt.Print(res.Stdout)
	}
	if res.Stderr != "" {
		fmt.Print(res.Stderr)
	}
	if res.Code != 0 {
		return fmt.Errorf("command failed for %s [%s]: exit %d", name, step, res.Code)
	}
	return nil
}
