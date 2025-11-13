package console

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/gopak/gopak-cli/internal/manager"
)

func (c *ConsoleUI) RunListImperative() error {
	groups := c.m.Tracked()
	keys := make([]manager.PackageKey, 0)
	for grp, names := range groups {
		for _, n := range names {
			keys = append(keys, manager.PackageKey{Source: grp, Name: n, Kind: kindOf(grp)})
		}
	}
	installed := c.m.GetVersionsInstalled(keys)
	fmt.Print(renderList(groups, installed))
	return nil
}

func renderList(groups map[string][]string, installed map[manager.PackageKey]string) string {
	var b strings.Builder
	srcs := make([]string, 0, len(groups))
	for k := range groups {
		srcs = append(srcs, k)
	}
	sort.Slice(srcs, func(i, j int) bool {
		if srcs[i] == "custom" {
			return false
		}
		if srcs[j] == "custom" {
			return true
		}
		return srcs[i] < srcs[j]
	})
	for _, grp := range srcs {
		b.WriteString(text.Bold.Sprint(grp) + "\n")
		tw := table.NewWriter()
		tw.SetStyle(table.StyleLight)
		tw.AppendHeader(table.Row{"Package", "Installed"})
		ns := append([]string{}, groups[grp]...)
		sort.Strings(ns)
		for _, name := range ns {
			k := manager.PackageKey{Source: grp, Name: name, Kind: kindOf(grp)}
			v := installed[k]
			if v == "" {
				v = "-"
			}
			tw.AppendRow(table.Row{name, v})
		}
		b.WriteString(tw.Render())
		b.WriteString("\n\n")
	}
	return b.String()
}
