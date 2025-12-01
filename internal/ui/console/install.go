package console

import (
	"fmt"
	"strings"
	"sync"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/gopak/gopak-cli/internal/manager"
)

type installEvent struct {
	k   manager.PackageKey
	ok  bool
	msg string
}

func (c *ConsoleUI) Install() error {
	groups := c.m.Tracked()
	status := map[manager.PackageKey]manager.VersionStatus{}
	lastLines := 0
	repaint := func(hideInstalled bool) {
		out := renderGroups(groups, status, hideInstalled)
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
			k := manager.PackageKey{Source: grp, Name: n, Kind: kindOf(grp)}
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
			keysAll = append(keysAll, manager.PackageKey{Source: grp, Name: n, Kind: kindOf(grp)})
		}
	}
	need := make([]manager.PackageKey, 0)
	for _, k := range keysAll {
		s := status[k]
		if s.Installed == "" {
			need = append(need, k)
		}
	}
	repaint(len(need) > 0)

	if len(need) == 0 {
		fmt.Println("Nothing to install")
		return nil
	}

	labels := make([]string, 0, len(need))
	labelByKey := map[string]manager.PackageKey{}
	for _, k := range need {
		s := status[k]
		var lbl string
		if s.Available != "" {
			lbl = fmt.Sprintf("%s/%s %s", k.Source, k.Name, s.Available)
		} else {
			lbl = fmt.Sprintf("%s/%s", k.Source, k.Name)
		}
		labels = append(labels, lbl)
		labelByKey[lbl] = k
	}

	selectedLabels := make([]string, 0)
	ms := &survey.MultiSelect{Message: "Select packages to install", Options: labels, Default: labels}
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
	if err := survey.AskOne(&survey.Confirm{Message: "Proceed to install selected?", Default: true}, &ok); err != nil {
		return err
	}
	if !ok {
		return nil
	}

	runner := manager.NewSudoRunner()
	defer runner.Close()

	var wgI sync.WaitGroup
	evCh := make(chan installEvent, 16)
	wgI.Add(1)
	go func() {
		defer wgI.Done()
		_ = c.m.InstallSelected(keysSelected, runner, func(k manager.PackageKey, ok bool, msg string) {
			evCh <- installEvent{k: k, ok: ok, msg: msg}
		})
	}()
	go func() { wgI.Wait(); close(evCh) }()

	for e := range evCh {
		if e.ok {
			action := e.msg
			if action == "" {
				action = "installed"
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
