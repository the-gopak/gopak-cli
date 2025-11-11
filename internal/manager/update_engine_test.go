package manager

import (
    "context"
    "reflect"
    "testing"

    "github.com/viktorprogger/universal-linux-installer/internal/config"
)

type mockReporter struct{
    inited bool
    groups map[string][]string
    installed map[PackageKey]string
    available map[PackageKey]string
    phases []string
    confirmed bool
    updates []PackageKey
    done bool
}

func (m *mockReporter) OnInit(groups map[string][]string){
    m.inited = true
    m.groups = map[string][]string{}
    for k,v := range groups { m.groups[k] = append([]string{}, v...) }
}
func (m *mockReporter) OnInstalled(k PackageKey, v string){
    if m.installed == nil { m.installed = map[PackageKey]string{} }
    m.installed[k] = v
}
func (m *mockReporter) OnAvailable(k PackageKey, v string){
    if m.available == nil { m.available = map[PackageKey]string{} }
    m.available[k] = v
}
func (m *mockReporter) OnPhaseDone(name string){ m.phases = append(m.phases, name) }
func (m *mockReporter) ConfirmProceed() bool { return m.confirmed }
func (m *mockReporter) OnUpdateStart() {}
func (m *mockReporter) OnPackageUpdated(k PackageKey, ok bool, errMsg string){ m.updates = append(m.updates, k) }
func (m *mockReporter) OnDone(){ m.done = true }

type mockRunner struct{ calls []string }
func (r *mockRunner) Run(name, step, script string, require *bool) error { r.calls = append(r.calls, name+":"+step); return nil }
func (r *mockRunner) Close() error { return nil }

func TestUpdateAll_Custom_NoProceed(t *testing.T){
    cfg := config.Config{ CustomPackages: []config.CustomPackage{{
        Name: "go",
        GetInstalledVersion: config.Command{Command: "echo 1.0.0"},
        GetLatestVersion: config.Command{Command: "echo 1.1.0"},
        Install: config.Command{Command: "echo install"},
    }}}
    m := New(cfg)
    rep := &mockReporter{confirmed:false}
    run := &mockRunner{}
    if err := m.UpdateAll(context.Background(), rep, run); err != nil { t.Fatalf("err: %v", err) }
    if !rep.inited || !rep.done { t.Fatalf("reporter not finished properly") }
    wantGroups := map[string][]string{"custom": {"go"}}
    if !reflect.DeepEqual(rep.groups, wantGroups) { t.Fatalf("groups: got=%v want=%v", rep.groups, wantGroups) }
    if len(run.calls) != 0 { t.Fatalf("runner should not be called when not confirmed, got %v", run.calls) }
}

func TestUpdateAll_Custom_Proceed(t *testing.T){
    cfg := config.Config{ CustomPackages: []config.CustomPackage{{
        Name: "tool",
        GetInstalledVersion: config.Command{Command: "echo 0.9.0"},
        GetLatestVersion: config.Command{Command: "echo 1.0.0"},
        Install: config.Command{Command: "echo install"},
    }}}
    m := New(cfg)
    rep := &mockReporter{confirmed:true}
    run := &mockRunner{}
    if err := m.UpdateAll(context.Background(), rep, run); err != nil { t.Fatalf("err: %v", err) }
    if len(run.calls) == 0 { t.Fatalf("runner should be called on proceed") }
    if len(rep.updates) == 0 || rep.updates[0].Name != "tool" { t.Fatalf("expected update event for tool, got %v", rep.updates) }
}
