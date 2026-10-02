package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/shimmerglass/i3-workspace-manager/history"
	"github.com/shimmerglass/i3-workspace-manager/i3"
	"github.com/shimmerglass/i3-workspace-manager/sel"
)

const (
	reconcileDelay = 500 * time.Millisecond
	closeTimeout   = time.Second
)

func (m *Manager) ActionDaemon() error {
	if err := m.reconcileDisplays(); err != nil {
		log.Printf("initial display reconcile failed: %s", err)
	}

	cmd := exec.Command("swaymsg", "-r", "-m", "-t", "subscribe", `["output"]`)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	lines := make(chan string)
	var scanErr error
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		scanErr = scanner.Err()
	}()

	debounce := time.NewTimer(reconcileDelay)
	debounce.Stop()
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return errors.Join(errors.New("output event stream closed"), scanErr, cmd.Wait())
			}
			log.Printf("output event: %s", line)
			debounce.Reset(reconcileDelay)
		case <-debounce.C:
			if err := m.reconcileDisplays(); err != nil {
				log.Printf("display reconcile failed: %s", err)
			}
		}
	}
}

func (m *Manager) ActionOpenProject(project string) error {
	projects, err := m.ProjectList()
	if err != nil {
		return err
	}
	if !slices.Contains(projects, project) {
		return fmt.Errorf("unknown project %q", project)
	}

	hist, err := history.Get()
	if err != nil {
		return err
	}
	return m.openAndRecord(hist, project)
}

func (m *Manager) ActionOpen() error {
	hist, err := history.Get()
	if err != nil {
		return err
	}

	projects, err := m.ProjectList()
	if err != nil {
		return err
	}
	hist.Sort(projects)

	project, err := sel.DoChoices(projects, "Open project", 0)
	if err != nil || project == "" {
		return err
	}
	return m.openAndRecord(hist, project)
}

func (m *Manager) ActionSelect() error {
	projects, err := m.OpenProjects()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return m.ActionOpen()
	}

	hist, err := history.Get()
	if err != nil {
		return err
	}
	hist.Sort(projects)

	current, _, err := m.CurrentProject()
	if err != nil {
		return err
	}
	pos := max(slices.Index(projects, current), 0)

	project, err := sel.DoChoices(projects, "Select project", pos)
	if err != nil || project == "" {
		return err
	}
	return m.openAndRecord(hist, project)
}

func (m *Manager) ActionHistoryGo(n int) error {
	current, ok, err := m.CurrentProject()
	if err != nil {
		return err
	}
	if ok {
		visible, err := m.IsProjectVisible(current)
		if err != nil {
			return err
		}
		if !visible {
			return m.OpenProject(current)
		}
	}

	hist, err := history.Get()
	if err != nil {
		return err
	}
	open, err := m.OpenProjects()
	if err != nil {
		return err
	}

	size := len(hist.Projects)
	pos := hist.Position
	for range size {
		pos = ((pos+n)%size + size) % size
		if !slices.Contains(open, hist.Projects[pos]) {
			continue
		}
		if err := m.OpenProject(hist.Projects[pos]); err != nil {
			return err
		}
		hist.Position = pos
		return history.Write(hist)
	}
	return nil
}

func (m *Manager) ActionClose() error {
	hist, err := history.Get()
	if err != nil {
		return err
	}

	open, err := m.OpenProjects()
	if err != nil || len(open) == 0 {
		return err
	}
	hist.Sort(open)

	projects, err := sel.DoChoicesMulti(open, "Close projects", len(open)-1)
	if err != nil {
		return err
	}

	wks, err := m.managedWorkspaces()
	if err != nil {
		return err
	}

	for _, project := range projects {
		if err := m.closeProject(wks, project); err != nil {
			return err
		}
		notify("i3wks", fmt.Sprintf("Closed project %s", project))
	}
	return nil
}

func (m *Manager) openAndRecord(hist *history.History, project string) error {
	if err := m.OpenProject(project); err != nil {
		return err
	}
	hist.Add(project)
	return history.Write(hist)
}

func (m *Manager) closeProject(wks []managedWks, project string) error {
	log.Printf("closing project %s", project)
	for _, w := range wks {
		if w.Project != project {
			continue
		}
		if err := m.closeWorkspace(w.Workspace); err != nil {
			return fmt.Errorf("closing project %s: %w", project, err)
		}
	}
	return nil
}

func (m *Manager) closeWorkspace(wk i3.Workspace) error {
	hasWindows, err := i3.WorkspaceHasWindows(m.sw, wk.Name)
	if err != nil || !hasWindows {
		return err
	}

	log.Printf("closing workspace %s", wk.Name)
	events, stop, err := i3.WinEvents()
	if err != nil {
		return err
	}
	defer stop()

	if err := i3.CloseWorkspace(m.sw, wk.Num); err != nil {
		return err
	}

	timeout := time.After(closeTimeout)
	for {
		select {
		case <-events:
		case <-timeout:
			hasWindows, err := i3.WorkspaceHasWindows(m.sw, wk.Name)
			if err != nil || !hasWindows {
				return err
			}
			return errors.Join(
				fmt.Errorf("some windows of workspace %s could not be closed", wk.Name),
				i3.SwitchToWorkspace(m.sw, wk.Name),
			)
		}

		hasWindows, err := i3.WorkspaceHasWindows(m.sw, wk.Name)
		if err != nil || !hasWindows {
			return err
		}
	}
}

func notify(title, msg string) {
	if err := beeep.Notify(title, msg, ""); err != nil {
		log.Printf("notification failed: %s", err)
	}
}
