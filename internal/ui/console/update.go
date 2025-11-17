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

// updateEvent captures a single package update outcome.
// Business: used for the update log to inform the user which items changed or failed.
type updateEvent struct {
	k   manager.PackageKey
	ok  bool
	msg string
}

func (c *ConsoleUI) Update() error {
	groups := c.m.Tracked()
	status := map[manager.PackageKey]manager.VersionStatus{}
	lastLines := 0
	repaint := func(hideUpToDate bool) {
		out := renderGroups(groups, status, hideUpToDate)
		if lastLines > 0 {
			fmt.Printf("\x1b[%dA", lastLines)
			fmt.Print("\x1b[J")
		}
		fmt.Print(out)
		lastLines = strings.Count(out, "\n")
	}
	repaint(false)

	updates := make(chan struct{}, 8)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for grp, names := range groups {
		grp := grp
		names := append([]string{}, names...)
		wg.Add(1)
		go func() {
			defer wg.Done()
			keys := make([]manager.PackageKey, 0, len(names))
			for _, n := range names {
				keys = append(keys, manager.PackageKey{Source: grp, Name: n, Kind: kindOf(grp)})
			}
			var ins, av map[manager.PackageKey]string
			var wg2 sync.WaitGroup
			wg2.Add(2)
			go func() { defer wg2.Done(); ins = c.m.GetVersionsInstalled(keys) }()
			go func() { defer wg2.Done(); av = c.m.GetVersionsAvailable(keys) }()
			wg2.Wait()
			mu.Lock()
			for k, v := range ins {
				s := status[k]
				s.Installed = v
				status[k] = s
			}
			for k, v := range av {
				s := status[k]
				s.Available = v
				status[k] = s
			}
			mu.Unlock()
			updates <- struct{}{}
		}()
	}
	go func() { wg.Wait(); close(updates) }()
	for range updates {
		repaint(false)
	}

	keysAll := make([]manager.PackageKey, 0)
	for grp, names := range groups {
		for _, n := range names {
			keysAll = append(keysAll, manager.PackageKey{Source: grp, Name: n, Kind: kindOf(grp)})
		}
	}
	need := make([]manager.PackageKey, 0)
	for _, k := range keysAll {
		s := status[k]
		if s.Installed == "" {
			need = append(need, k)
			continue
		}
		if s.Available != "" && s.Installed != s.Available {
			need = append(need, k)
		}
	}
	repaint(len(need) > 0)

	if len(need) == 0 {
		fmt.Println("Nothing to update")
		return nil
	}

	labels := make([]string, 0, len(need))
	labelByKey := map[string]manager.PackageKey{}
	for _, k := range need {
		s := status[k]
		var lbl string
		if s.Installed == "" {
			if s.Available != "" {
				lbl = fmt.Sprintf("%s/%s install %s", k.Source, k.Name, s.Available)
			} else {
				lbl = fmt.Sprintf("%s/%s install", k.Source, k.Name)
			}
		} else {
			lbl = fmt.Sprintf("%s/%s %s -> %s", k.Source, k.Name, s.Installed, s.Available)
		}
		labels = append(labels, lbl)
		labelByKey[lbl] = k
	}

	selectedLabels := make([]string, 0)
	ms := &survey.MultiSelect{Message: "Select packages to update", Options: labels, Default: labels}
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
	if err := survey.AskOne(&survey.Confirm{Message: "Proceed to update selected?", Default: true}, &ok); err != nil {
		return err
	}
	if !ok {
		return nil
	}

	runner := manager.NewSudoRunner()
	defer runner.Close()

	var wgU sync.WaitGroup
	evCh := make(chan updateEvent, 16)
	wgU.Add(1)
	go func() {
		defer wgU.Done()
		_ = c.m.UpdateSelected(keysSelected, runner, func(k manager.PackageKey, ok bool, msg string) {
			evCh <- updateEvent{k: k, ok: ok, msg: msg}
		})
	}()
	go func() { wgU.Wait(); close(evCh) }()

	for e := range evCh {
		if e.ok {
			action := e.msg
			if action == "" {
				action = "updated"
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
			k := manager.PackageKey{Source: grp, Name: name, Kind: kindOf(grp)}
			s := status[k]
			cur := ""
			ins := ""
			if s.Installed != "" || s.Available != "" {
				if s.Available == "" {
					cur = s.Installed
				} else if s.Installed == "" {
					ins = s.Available
				} else if s.Installed == s.Available {
					if hideUpToDate {
						continue
					}
					cur = text.FgHiBlack.Sprint(s.Installed)
					ins = text.FgHiBlack.Sprint(s.Available)
				} else {
					cur = colorGreen(s.Installed)
					ins = colorGreen(s.Available)
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
