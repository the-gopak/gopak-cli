package manager

import "testing"

func TestTopoOrder(t *testing.T) {
	nodes := map[string][]string{
		"a": {},
		"b": {"a"},
		"c": {"a", "b"},
	}
	ord, ok := topoOrder(nodes)
	if !ok {
		t.Fatalf("cycle detected")
	}
	pos := map[string]int{}
	for i, n := range ord {
		pos[n] = i
	}
	if !(pos["a"] < pos["b"] && pos["b"] < pos["c"]) {
		t.Fatalf("wrong order: %v", ord)
	}
}

func TestTopoOrderCycle(t *testing.T) {
	nodes := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	_, ok := topoOrder(nodes)
	if ok {
		t.Fatalf("expected cycle")
	}
}
