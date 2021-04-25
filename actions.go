package main

import (
	"fmt"

	"github.com/gen2brain/beeep"
	"github.com/shimmerglass/i3-workspace-manager/history"
	"github.com/shimmerglass/i3-workspace-manager/i3"
	"github.com/shimmerglass/i3-workspace-manager/sel"
)

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
	if err != nil {
		return err
	}
	if project == "" {
		return nil
	}

	err = m.OpenProject(project)
	if err != nil {
		return err
	}

	hist.Add(project)

	return history.Write(hist)
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

	current, ok, err := m.CurrentProject()
	if err != nil {
		return err
	}

	pos := 0
	if ok {
		for i, p := range projects {
			if p == current {
				pos = i
			}
		}
	}

	project, err := sel.DoChoices(projects, "Select project", pos)
	if err != nil {
		return err
	}
	if project == "" {
		return nil
	}

	err = m.OpenProject(project)
	if err != nil {
		return err
	}

	hist.Add(project)

	return history.Write(hist)
}

func (m *Manager) ActionHistoryGo(n int) error {
	current, ok, err := m.CurrentProject()
	if err != nil {
		return err
	}
	if ok {
		visible, err := m.IsProjectVisble(current)
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

	openProjects, err := m.OpenProjects()
	if err != nil {
		return err
	}
	if len(openProjects) == 0 {
		return nil
	}

	selected := ""
	pos := hist.Position
Outer:
	for i := 0; i < len(hist.Projects); i++ {
		pos = (pos + n) % len(hist.Projects)
		if pos < 0 {
			pos = len(hist.Projects) - 1
		}

		for _, o := range openProjects {
			if o == hist.Projects[pos] {
				selected = o
				break Outer
			}
		}
	}

	if selected == "" {
		return nil
	}

	err = m.OpenProject(selected)
	if err != nil {
		return err
	}

	hist.Position = pos

	return history.Write(hist)
}

func (m *Manager) ActionClose() error {
	hist, err := history.Get()
	if err != nil {
		return err
	}

	opened, err := m.OpenProjects()
	if err != nil {
		return err
	}
	if len(opened) == 0 {
		return nil
	}

	hist.Sort(opened)

	project, err := sel.DoChoices(opened, "Close project", len(opened)-1)
	if err != nil {
		return err
	}
	if project == "" {
		return nil
	}

	workspaces, err := i3.Workspaces()
	if err != nil {
		return err
	}

	clean := true
	for _, wk := range workspaces {
		p, ok := workspaceProject(wk)
		if !ok || p != project {
			continue
		}

		err := i3.CloseWorkspace(wk.Num)
		if err != nil {
			return err
		}

		if i3.WorkspaceHasWindows(wk.Name) {
			clean = false
		}
	}

	if clean {
		beeep.Notify("i3wks", fmt.Sprintf("Closed project %s", project), "")
	} else {
		beeep.Notify("i3wks", fmt.Sprintf("Some windows could not be closed closing project %s", project), "")
	}

	return nil
}
