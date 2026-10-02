package main

import (
	"testing"

	"github.com/joshuarubin/go-sway"
	"github.com/shimmerglass/i3-workspace-manager/i3"
)

func TestWorkspaceNameRoundTrip(t *testing.T) {
	info := wksInfo{Project: "foo/bar baz", Display: "left"}
	name := formatWorkspaceName(21, "label", info)

	got, ok, err := parseWorkspaceName(name)
	if err != nil || !ok {
		t.Fatalf("parse %q: ok=%v err=%v", name, ok, err)
	}
	if got != info {
		t.Fatalf("got %+v, want %+v", got, info)
	}
}

func TestParseWorkspaceNameUnmanaged(t *testing.T) {
	for _, name := range []string{"1", "20:code", "a​b​c"} {
		if _, ok, err := parseWorkspaceName(name); ok || err != nil {
			t.Errorf("%q: ok=%v err=%v, want unmanaged", name, ok, err)
		}
	}
}

func TestParseWorkspaceNameGarbage(t *testing.T) {
	for _, meta := range []string{"x", "‌‌‌‌‌‌‌‌"} {
		if _, ok, err := parseWorkspaceName("label​" + meta); ok || err == nil {
			t.Errorf("%q: ok=%v err=%v, want error", meta, ok, err)
		}
	}
}

func outputs(names ...string) []i3.Output {
	var res []i3.Output
	for i, n := range names {
		res = append(res, i3.Output{Name: n, Active: true, Rect: sway.Rect{X: int64(i) * 1920}})
	}
	return res
}

func TestResolveRole(t *testing.T) {
	unsorted := append(outputs("DP-1", "DP-2"), i3.Output{Name: "HDMI-1", Active: false})
	unsorted[0].Rect.X, unsorted[1].Rect.X = 1920, 0
	active := activeOutputsSorted(unsorted)

	tests := map[string]string{
		"left":   "DP-2",
		"right":  "DP-1",
		"DP-1":   "DP-1",
		"HDMI-1": "",
		"middle": "",
	}
	for role, want := range tests {
		if got := resolveRole(active, role); got != want {
			t.Errorf("resolveRole(%q) = %q, want %q", role, got, want)
		}
	}

	if got := resolveRole(nil, "left"); got != "" {
		t.Errorf("no outputs: got %q", got)
	}
}

func mwks(project, output string, visible, focused bool) managedWks {
	return managedWks{
		Workspace: i3.Workspace{Output: output, Visible: visible, Focused: focused},
		wksInfo:   wksInfo{Project: project},
	}
}

func TestProjectVisible(t *testing.T) {
	tests := []struct {
		name string
		wks  []managedWks
		want bool
	}{
		{"all shown", []managedWks{mwks("a", "L", true, false), mwks("a", "R", true, false)}, true},
		{"one hidden", []managedWks{mwks("a", "L", true, false), mwks("a", "R", false, false), mwks("b", "R", true, false)}, false},
		{"stacked on one output", []managedWks{mwks("a", "L", true, false), mwks("a", "L", false, false)}, true},
		{"missing workspace", []managedWks{mwks("a", "L", true, false)}, false},
		{"absent", []managedWks{mwks("b", "L", true, false), mwks("b", "R", true, false)}, false},
	}
	for _, tt := range tests {
		if got := projectVisible(tt.wks, "a", 2); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestCurrentProjectPrefersFocused(t *testing.T) {
	wks := []managedWks{mwks("a", "L", true, false), mwks("b", "R", true, true)}
	if got, ok := currentProject(wks); !ok || got != "b" {
		t.Fatalf("got %q %v, want b", got, ok)
	}
}
