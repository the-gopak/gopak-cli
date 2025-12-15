package console

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/gopak/gopak-cli/internal/manager"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type packageEvent struct {
	k   manager.PackageKey
	ok  bool
	msg string
}

type filterFunc func(status manager.VersionStatus) bool

type labelFunc func(k manager.PackageKey, status manager.VersionStatus) string

func versionsEqual(installed, available string) bool {
	if installed == "" || available == "" {
		return installed == available
	}
	ni := manager.NormalizeVersion(installed)
	na := manager.NormalizeVersion(available)
	if ni == "" || na == "" {
		return installed == available
	}
	return manager.CompareVersions(na, ni) == 0
}

func versionsNeedUpdate(installed, available string) bool {
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

func displayVersion(v string) string {
	n := manager.NormalizeVersion(v)
	if n != "" {
		return n
	}
	return v
}

func filterForUpdate(s manager.VersionStatus) bool {
	return versionsNeedUpdate(s.Installed, s.Available)
}

func filterForInstall(s manager.VersionStatus) bool {
	return s.Installed == ""
}

func labelForUpdate(k manager.PackageKey, s manager.VersionStatus) string {
	return fmt.Sprintf("%s/%s %s -> %s", k.Source, k.Name, displayVersion(s.Installed), displayVersion(s.Available))
}

func labelForInstall(k manager.PackageKey, s manager.VersionStatus) string {
	if s.Available != "" {
		return fmt.Sprintf("%s/%s %s", k.Source, k.Name, displayVersion(s.Available))
	}
	return fmt.Sprintf("%s/%s", k.Source, k.Name)
}

func (c *ConsoleUI) packageSelectionFlow(
	op manager.Operation,
	filter filterFunc,
	labelFn labelFunc,
	emptyMsg string,
	selectMsg string,
	confirmMsg string,
) error {
	groups := c.m.Tracked()
	status := map[manager.PackageKey]manager.VersionStatus{}
	lastLines := 0
	repaint := func(hideCompleted bool) {
		out := renderGroups(groups, status, hideCompleted)
		if lastLines > 0 {
			fmt.Printf("\x1b[%dA", lastLines)
			fmt.Print("\x1b[J")
		}
		fmt.Print(out)
		lastLines = strings.Count(out, "\n")
	}
	repaint(false)

	updates := make(chan struct{}, 32)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for grp, names := range groups {
		for _, n := range names {
			k := manager.PackageKey{Source: grp, Name: n, Kind: manager.KindOf(grp)}
			if !c.m.HasCommand(k, op) {
				continue
			}
			wg.Add(1)
			go func(k manager.PackageKey) {
				defer wg.Done()
				ins := c.m.GetVersionInstalled(k)
				mu.Lock()
				s := status[k]
				s.Installed = ins
				status[k] = s
				mu.Unlock()
				updates <- struct{}{}

				av := c.m.GetVersionAvailable(k)
				mu.Lock()
				s = status[k]
				s.Available = av
				status[k] = s
				mu.Unlock()
				updates <- struct{}{}
			}(k)
		}
	}
	go func() { wg.Wait(); close(updates) }()
	for range updates {
		repaint(false)
	}

	keysAll := make([]manager.PackageKey, 0)
	for grp, names := range groups {
		for _, n := range names {
			k := manager.PackageKey{Source: grp, Name: n, Kind: manager.KindOf(grp)}
			if c.m.HasCommand(k, op) {
				keysAll = append(keysAll, k)
			}
		}
	}
	need := make([]manager.PackageKey, 0)
	for _, k := range keysAll {
		if filter(status[k]) {
			need = append(need, k)
		}
	}
	repaint(len(need) > 0)

	if len(need) == 0 {
		fmt.Println(emptyMsg)
		return nil
	}

	labels := make([]string, 0, len(need))
	labelByKey := map[string]manager.PackageKey{}
	for _, k := range need {
		lbl := labelFn(k, status[k])
		labels = append(labels, lbl)
		labelByKey[lbl] = k
	}

	selectedLabels := make([]string, 0)
	ms := &survey.MultiSelect{Message: selectMsg, Options: labels, Default: labels}
	if err := survey.AskOne(ms, &selectedLabels); err != nil {
		return err
	}
	if len(selectedLabels) == 0 {
		fmt.Println("Nothing selected")
		return nil
	}
	keysSelected := make([]manager.PackageKey, 0, len(selectedLabels))
	for _, l := range selectedLabels {
		keysSelected = append(keysSelected, labelByKey[l])
	}
	fmt.Printf("%v\n", keysSelected)

	ok := false
	if err := survey.AskOne(&survey.Confirm{Message: confirmMsg, Default: true}, &ok); err != nil {
		return err
	}
	if !ok {
		return nil
	}

	runner := manager.NewSudoRunner()
	defer runner.Close()

	var wgE sync.WaitGroup
	evCh := make(chan packageEvent, 16)
	wgE.Add(1)
	go func() {
		defer wgE.Done()
		_ = c.m.ExecuteSelected(keysSelected, op, runner, func(k manager.PackageKey, ok bool, msg string) {
			evCh <- packageEvent{k: k, ok: ok, msg: msg}
		})
	}()
	go func() { wgE.Wait(); close(evCh) }()

	for e := range evCh {
		if e.ok {
			action := e.msg
			if action == "" {
				action = "updated"
				if op == manager.OpInstall {
					action = "installed"
				}
			}
			fmt.Println(colorGreen(action + ": " + e.k.Name))
		} else {
			fmt.Println(colorRed("failed:  " + e.k.Name))
			if e.msg != "" {
				fmt.Println(e.msg)
			}
		}
	}
	return nil
}

func (c *ConsoleUI) Update() error {
	return c.packageSelectionFlow(
		manager.OpUpdate,
		filterForUpdate,
		labelForUpdate,
		"Nothing to update",
		"Select packages to update",
		"Proceed to update selected?",
	)
}

func (c *ConsoleUI) Install() error {
	return c.packageSelectionFlow(
		manager.OpInstall,
		filterForInstall,
		labelForInstall,
		"Nothing to install",
		"Select packages to install",
		"Proceed to install selected?",
	)
}

func renderGroups(groups map[string][]string, status map[manager.PackageKey]manager.VersionStatus, hideUpToDate bool) string {
	var b strings.Builder
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i] == "custom" {
			return false
		}
		if keys[j] == "custom" {
			return true
		}
		return keys[i] < keys[j]
	})
	for _, grp := range keys {
		tw := table.NewWriter()
		tw.SetStyle(table.StyleLight)
		tw.AppendHeader(table.Row{"Package", "Current", "Installable"})
		names := append([]string{}, groups[grp]...)
		sort.Strings(names)
		for _, name := range names {
			k := manager.PackageKey{Source: grp, Name: name, Kind: manager.KindOf(grp)}
			s := status[k]
			cur := ""
			ins := ""
			if s.Installed != "" || s.Available != "" {
				if s.Available == "" {
					cur = displayVersion(s.Installed)
				} else if s.Installed == "" {
					ins = displayVersion(s.Available)
				} else if versionsEqual(s.Installed, s.Available) {
					if hideUpToDate {
						continue
					}
					cur = text.FgHiBlack.Sprint(displayVersion(s.Installed))
					ins = text.FgHiBlack.Sprint(displayVersion(s.Available))
				} else {
					cur = colorGreen(displayVersion(s.Installed))
					ins = colorGreen(displayVersion(s.Available))
				}
			}
			tw.AppendRow(table.Row{name, cur, ins})
		}
		if tw.Length() > 0 {
			b.WriteString(text.Bold.Sprint(grp) + "\n")
			b.WriteString(tw.Render())
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

func colorGreen(s string) string { return text.FgGreen.Sprint(s) }
func colorRed(s string) string   { return text.FgRed.Sprint(s) }
