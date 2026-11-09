package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

func TestNewOutput(t *testing.T) {
	before := []i3.Output{{Name: "DP-1"}, {Name: "HDMI-A-1"}}
	after := []i3.Output{{Name: "DP-1"}, {Name: "HEADLESS-3"}, {Name: "HDMI-A-1"}}

	if got, ok := newOutput(before, after); !ok || got != "HEADLESS-3" {
		t.Fatalf("newOutput = %q, %v", got, ok)
	}
	if _, ok := newOutput(before, before); ok {
		t.Fatal("newOutput found an output that was already there")
	}
}

func TestRestoreWorkspacesCommandNoop(t *testing.T) {
	wks := []i3.Workspace{
		{Name: "1", Output: "DP-1", Visible: true, Focused: true},
		{Name: "2", Output: "DP-1"},
	}
	if got := restoreWorkspacesCommand(wks, wks); got != "" {
		t.Fatalf("got %q, want no command", got)
	}
	if got := restoreWorkspacesCommand(wks, wks[:1]); got != "" {
		t.Fatalf("got %q for a destroyed workspace, want no command", got)
	}
}

func TestRestoreWorkspacesCommandMovesBack(t *testing.T) {
	before := []i3.Workspace{
		{Name: "1", Output: "DP-1", Visible: true, Focused: true},
		{Name: "2", Output: "HDMI-A-1", Visible: true},
		{Name: "21:x", Output: "HDMI-A-1"},
	}
	after := []i3.Workspace{
		{Name: "1", Output: "DP-1", Visible: true, Focused: true},
		{Name: "2", Output: "HDMI-A-1", Visible: true},
		{Name: "21:x", Output: "DP-1"},
	}
	want := `workspace --no-auto-back-and-forth "21:x"; move workspace to output "HDMI-A-1"; ` +
		`workspace --no-auto-back-and-forth "2"; workspace --no-auto-back-and-forth "1"`
	if got := restoreWorkspacesCommand(before, after); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestHasChild(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	if !hasChild(os.Getpid()) {
		t.Fatal("test process has a child but hasChild says no")
	}
	if hasChild(cmd.Process.Pid) {
		t.Fatal("sleep has no child but hasChild says yes")
	}
}
