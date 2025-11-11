package console

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/viktorprogger/universal-linux-installer/internal/manager"
)

type consoleReporter struct {
	mu        sync.Mutex
	groups    map[string][]string
	status    map[manager.PackageKey]manager.VersionStatus
	last      time.Time
	throttle  time.Duration
}

func NewConsoleReporter() manager.UpdateReporter {
	return &consoleReporter{status: map[manager.PackageKey]manager.VersionStatus{}, throttle: 50 * time.Millisecond}
}

func (r *consoleReporter) OnInit(groups map[string][]string) {
	r.mu.Lock()
	r.groups = map[string][]string{}
	for k, v := range groups { r.groups[k] = append([]string{}, v...) }
	r.mu.Unlock()
	r.render(true)
}

func (r *consoleReporter) OnInstalled(k manager.PackageKey, version string) {
	r.mu.Lock()
	s := r.status[k]
	s.Installed = version
	r.status[k] = s
	r.mu.Unlock()
	r.render(false)
}

func (r *consoleReporter) OnAvailable(k manager.PackageKey, version string) {
	r.mu.Lock()
	s := r.status[k]
	s.Available = version
	r.status[k] = s
	r.mu.Unlock()
	r.render(false)
}

func (r *consoleReporter) OnPhaseDone(name string) { r.render(true) }

func (r *consoleReporter) ConfirmProceed() bool {
	r.render(true)
	fmt.Println()
	fmt.Print("Proceed to update all? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" && (line == "n" || line == "N") { return false }
	return true
}

func (r *consoleReporter) OnUpdateStart() {}

func (r *consoleReporter) OnPackageUpdated(k manager.PackageKey, ok bool, errMsg string) {
	if ok {
		fmt.Println("\x1b[32mupdated: " + k.Name + "\x1b[0m")
	} else {
		fmt.Println("\x1b[31mfailed:  " + k.Name + "\x1b[0m")
		if errMsg != "" { fmt.Println(errMsg) }
	}
}

func (r *consoleReporter) OnDone() { r.render(true) }

func (r *consoleReporter) render(force bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !force && time.Since(r.last) < r.throttle { return }
	r.last = time.Now()
	fmt.Print("\033[H\033[2J")
	keys := make([]string, 0, len(r.groups))
	for k := range r.groups { keys = append(keys, k) }
	sort.Slice(keys, func(i, j int) bool {
		if keys[i] == "custom" { return false }
		if keys[j] == "custom" { return true }
		return keys[i] < keys[j]
	})
	for _, grp := range keys {
		fmt.Printf("[%s]\n", grp)
		names := append([]string{}, r.groups[grp]...)
		sort.Strings(names)
		for _, n := range names {
			k := manager.PackageKey{Source: grp, Name: n, Kind: kindOf(grp)}
			s := r.status[k]
			line := fmt.Sprintf("  %s:", n)
			if s.Installed != "" || s.Available != "" {
				if s.Available == "" {
					line = fmt.Sprintf("  %s: %s", n, s.Installed)
				} else if s.Installed == "" {
					line = fmt.Sprintf("  %s: -> %s", n, s.Available)
				} else if s.Installed == s.Available {
					line = fmt.Sprintf("  %s: %s", n, s.Installed)
					fmt.Println("\x1b[90m" + line + "\x1b[0m")
					continue
				} else {
					line = fmt.Sprintf("  %s: %s -> %s", n, s.Installed, s.Available)
					fmt.Println("\x1b[32m" + line + "\x1b[0m")
					continue
				}
			}
			fmt.Println(line)
		}
		fmt.Println()
	}
}

func kindOf(group string) string { if group == "custom" { return "custom" } ; return "source" }
