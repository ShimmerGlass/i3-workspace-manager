package main

import (
	"github.com/shimmerglass/i3-workspace-manager/history"
	"github.com/shimmerglass/i3-workspace-manager/i3"
	"github.com/shimmerglass/i3-workspace-manager/sel"
)

func (m *Manager) ActionOpen() error {
	project, err := m.projectList()
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

	err = history.SetIndex(0)
	if err != nil {
		return err
	}
	currents, err := m.OpenProjects()
	if err != nil {
		return err
	}
	err = history.Filter(currents)
	if err != nil {
		return err
	}

	return history.Add(project)
}

func (m *Manager) ActionSelect() error {
	projects, err := m.OpenProjects()
	if err != nil {
		return err
	}

	project, err := sel.DoChoices(projects)
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

	err = history.SetIndex(0)
	if err != nil {
		return err
	}
	currents, err := m.OpenProjects()
	if err != nil {
		return err
	}
	err = history.Filter(currents)
	if err != nil {
		return err
	}

	return history.Add(project)
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
	if len(hist) == 0 {
		return nil
	}

	histIndex, err := history.Index()
	if err != nil {
		return err
	}

	histIndex += n
	if histIndex < 0 {
		histIndex = 0
	}
	if histIndex >= len(hist) {
		histIndex = len(hist) - 1
	}

	err = m.OpenProject(hist[histIndex])
	if err != nil {
		return err
	}

	err = history.SetIndex(histIndex)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) ActionClose() error {
	current, ok, err := m.CurrentProject()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	wks, err := m.ProjectWks(current)
	if err != nil {
		return err
	}

	for _, w := range wks {
		if w != nil {
			err := i3.CloseWorkspace(w.Num)
			if err != nil {
				return err
			}
		}
	}

	currents, err := m.OpenProjects()
	if err != nil {
		return err
	}
	err = history.Filter(currents)
	if err != nil {
		return err
	}

	return m.ActionHistoryGo(1)
}
