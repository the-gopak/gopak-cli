package console

import (
    "fmt"
    "sort"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/viktorprogger/universal-linux-installer/internal/manager"
)

type initMsg struct{ groups map[string][]string }
type installedMsg struct{ k manager.PackageKey; version string }
type availableMsg struct{ k manager.PackageKey; version string }
type phaseDoneMsg struct{ name string }
type updateStartMsg struct{}
type pkgUpdatedMsg struct{ k manager.PackageKey; ok bool; err string }
type confirmRequestMsg struct{ ch chan bool }
type finishMsg struct{}

type model struct {
    groups       map[string][]string
    status       map[manager.PackageKey]manager.VersionStatus
    logs         []string
    confirmCh    chan bool
    confirmStage bool
    ready        chan struct{}
}

func (m *model) Init() tea.Cmd {
    if m.ready != nil { close(m.ready); m.ready = nil }
    return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case initMsg:
        m.groups = map[string][]string{}
        for k, v := range msg.groups { m.groups[k] = append([]string{}, v...) }
        return m, nil
    case installedMsg:
        s := m.status[msg.k]
        s.Installed = msg.version
        if m.status == nil { m.status = map[manager.PackageKey]manager.VersionStatus{} }
        m.status[msg.k] = s
        return m, nil
    case availableMsg:
        s := m.status[msg.k]
        s.Available = msg.version
        if m.status == nil { m.status = map[manager.PackageKey]manager.VersionStatus{} }
        m.status[msg.k] = s
        return m, nil
    case updateStartMsg:
        return m, nil
    case pkgUpdatedMsg:
        if msg.ok {
            m.logs = append(m.logs, "\x1b[32mupdated: "+msg.k.Name+"\x1b[0m")
        } else {
            m.logs = append(m.logs, "\x1b[31mfailed:  "+msg.k.Name+"\x1b[0m")
            if msg.err != "" { m.logs = append(m.logs, msg.err) }
        }
        return m, nil
    case confirmRequestMsg:
        m.confirmCh = msg.ch
        m.confirmStage = true
        return m, nil
    case tea.KeyMsg:
        if m.confirmStage {
            k := msg.String()
            if k == "y" || k == "Y" || k == "enter" {
                if m.confirmCh != nil { m.confirmCh <- true }
                m.confirmCh = nil
                m.confirmStage = false
                return m, nil
            }
            if k == "n" || k == "N" {
                if m.confirmCh != nil { m.confirmCh <- false }
                m.confirmCh = nil
                m.confirmStage = false
                return m, nil
            }
        }
        return m, nil
    case phaseDoneMsg:
        return m, nil
    case finishMsg:
        return m, tea.Quit
    default:
        return m, nil
    }
}

func (m *model) View() string {
    var b strings.Builder
    keys := make([]string, 0, len(m.groups))
    for k := range m.groups { keys = append(keys, k) }
    sort.Slice(keys, func(i, j int) bool {
        if keys[i] == "custom" { return false }
        if keys[j] == "custom" { return true }
        return keys[i] < keys[j]
    })
    for _, grp := range keys {
        b.WriteString("["+grp+"]\n")
        names := append([]string{}, m.groups[grp]...)
        sort.Strings(names)
        for _, n := range names {
            k := manager.PackageKey{Source: grp, Name: n, Kind: kindOf(grp)}
            s := m.status[k]
            line := fmt.Sprintf("  %s:", n)
            if s.Installed != "" || s.Available != "" {
                if s.Available == "" {
                    line = fmt.Sprintf("  %s: %s", n, s.Installed)
                } else if s.Installed == "" {
                    line = fmt.Sprintf("  %s: -> %s", n, s.Available)
                } else if s.Installed == s.Available {
                    line = fmt.Sprintf("  %s: %s", n, s.Installed)
                    b.WriteString("\x1b[90m" + line + "\x1b[0m\n")
                    continue
                } else {
                    line = fmt.Sprintf("  %s: %s -> %s", n, s.Installed, s.Available)
                    b.WriteString("\x1b[32m" + line + "\x1b[0m\n")
                    continue
                }
            }
            b.WriteString(line+"\n")
        }
        b.WriteString("\n")
    }
    if m.confirmStage {
        b.WriteString("\nProceed to update all? [Y/n]: ")
    }
    if len(m.logs) > 0 {
        b.WriteString("\n")
        for _, l := range m.logs { b.WriteString(l+"\n") }
    }
    return b.String()
}

type bubbleReporter struct{ p *tea.Program }

func NewConsoleReporter() manager.UpdateReporter { return NewConsoleReporterWithOptions() }

func NewConsoleReporterWithOptions(opts ...tea.ProgramOption) manager.UpdateReporter {
    ready := make(chan struct{})
    m := &model{groups: map[string][]string{}, status: map[manager.PackageKey]manager.VersionStatus{}, ready: ready}
    p := tea.NewProgram(m, opts...)
    go func() { _, _ = p.Run() }()
    <-ready
    return &bubbleReporter{p: p}
}

func (r *bubbleReporter) OnInit(groups map[string][]string) {
    cp := map[string][]string{}
    for k, v := range groups { cp[k] = append([]string{}, v...) }
    r.p.Send(initMsg{groups: cp})
}

func (r *bubbleReporter) OnInstalled(k manager.PackageKey, version string) { r.p.Send(installedMsg{k: k, version: version}) }

func (r *bubbleReporter) OnAvailable(k manager.PackageKey, version string) { r.p.Send(availableMsg{k: k, version: version}) }

func (r *bubbleReporter) OnPhaseDone(name string) { r.p.Send(phaseDoneMsg{name: name}) }

func (r *bubbleReporter) ConfirmProceed() bool {
    ch := make(chan bool, 1)
    r.p.Send(confirmRequestMsg{ch: ch})
    res := <-ch
    return res
}

func (r *bubbleReporter) OnUpdateStart() { r.p.Send(updateStartMsg{}) }

func (r *bubbleReporter) OnPackageUpdated(k manager.PackageKey, ok bool, errMsg string) {
    r.p.Send(pkgUpdatedMsg{k: k, ok: ok, err: errMsg})
}

func (r *bubbleReporter) OnDone() { r.p.Send(finishMsg{}) }

func kindOf(group string) string { if group == "custom" { return "custom" } ; return "source" }
