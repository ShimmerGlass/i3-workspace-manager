package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

const (
	wakeGrace   = time.Second
	wakeTimeout = 2 * time.Second
)

func (m *Manager) wakeIfStuck(conID int64, wksName, wksOutput string) error {
	wks, err := i3.Workspaces(m.sw)
	if err != nil {
		return err
	}
	if slices.ContainsFunc(wks, func(w i3.Workspace) bool { return w.Name == wksName && w.Visible }) {
		return nil
	}
	tree, err := i3.Tree(m.sw)
	if err != nil {
		return err
	}
	n := i3.FindContainer(tree, conID)
	if n == nil || n.PID == nil {
		return nil
	}
	pid := int(*n.PID)
	if waitChild(pid, wakeGrace) {
		return nil
	}

	log.Printf("window %d on hidden workspace %s started no process, waking it on a headless output", conID, wksName)
	return m.wakeOnHeadless(conID, pid, wksName, wksOutput)
}

func waitChild(pid int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for {
		if hasChild(pid) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func hasChild(pid int) bool {
	files, _ := filepath.Glob(fmt.Sprintf("/proc/%d/task/*/children", pid))
	for _, f := range files {
		if b, err := os.ReadFile(f); err == nil && len(bytes.TrimSpace(b)) > 0 {
			return true
		}
	}
	return false
}

func (m *Manager) wakeOnHeadless(conID int64, pid int, wksName, wksOutput string) (err error) {
	outsBefore, err := i3.Outputs(m.sw)
	if err != nil {
		return err
	}
	wksBefore, err := i3.Workspaces(m.sw)
	if err != nil {
		return err
	}
	focused := focusedWorkspace(wksBefore)
	if focused == "" {
		return errors.New("no focused workspace")
	}

	if err := i3.RunCommand(m.sw, "create_output; workspace --no-auto-back-and-forth "+i3.Quote(focused)); err != nil {
		return fmt.Errorf("creating headless output: %w", err)
	}
	outsAfter, err := i3.Outputs(m.sw)
	if err != nil {
		return err
	}
	headless, ok := newOutput(outsBefore, outsAfter)
	if !ok {
		return errors.New("create_output did not add an output")
	}
	defer func() {
		err = errors.Join(err, m.unplugHeadless(headless, wksBefore))
	}()

	wks, err := i3.Workspaces(m.sw)
	if err != nil {
		return err
	}
	target := ""
	for _, w := range wks {
		if w.Output == headless {
			target = w.Name
			break
		}
	}
	if target == "" {
		return fmt.Errorf("no workspace on headless output %s", headless)
	}

	if err := i3.MoveContainerToWorkspace(m.sw, conID, target); err != nil {
		return err
	}
	if !waitChild(pid, wakeTimeout) {
		log.Printf("window %d still started no process after %s on %s", conID, wakeTimeout, headless)
	}

	back := fmt.Sprintf("[con_id=%d] move container to workspace %s", conID, i3.Quote(wksName))
	if wksOutput != "" {
		back = "workspace " + i3.Quote(wksName) + " output " + i3.Quote(wksOutput) + "; " + back
	}
	return i3.RunCommand(m.sw, back)
}

func (m *Manager) unplugHeadless(headless string, wksBefore []i3.Workspace) error {
	if err := i3.RunCommand(m.sw, "output "+i3.Quote(headless)+" unplug"); err != nil {
		return fmt.Errorf("unplugging headless output %s: %w", headless, err)
	}
	wksAfter, err := i3.Workspaces(m.sw)
	if err != nil {
		return err
	}
	cmd := restoreWorkspacesCommand(wksBefore, wksAfter)
	if cmd == "" {
		return nil
	}
	log.Printf("workspaces moved by the headless output, putting them back")
	return i3.RunCommand(m.sw, cmd)
}

func focusedWorkspace(wks []i3.Workspace) string {
	for _, w := range wks {
		if w.Focused {
			return w.Name
		}
	}
	return ""
}

func newOutput(before, after []i3.Output) (string, bool) {
	for _, o := range after {
		if !slices.ContainsFunc(before, func(b i3.Output) bool { return b.Name == o.Name }) {
			return o.Name, true
		}
	}
	return "", false
}

func restoreWorkspacesCommand(before, after []i3.Workspace) string {
	var parts []string
	for _, b := range before {
		i := slices.IndexFunc(after, func(a i3.Workspace) bool { return a.Name == b.Name })
		if i < 0 || after[i].Output == b.Output {
			continue
		}
		parts = append(parts, "workspace --no-auto-back-and-forth "+i3.Quote(b.Name), "move workspace to output "+i3.Quote(b.Output))
	}
	if len(parts) == 0 {
		return ""
	}
	var focused string
	for _, b := range before {
		switch {
		case b.Focused:
			focused = b.Name
		case b.Visible:
			parts = append(parts, "workspace --no-auto-back-and-forth "+i3.Quote(b.Name))
		}
	}
	if focused != "" {
		parts = append(parts, "workspace --no-auto-back-and-forth "+i3.Quote(focused))
	}
	return strings.Join(parts, "; ")
}
