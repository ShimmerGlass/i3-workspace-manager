package main

import (
	"sort"
	"strings"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

const (
	workspaceStart = 20
)

func workspaceProject(wk i3.Workspace) (string, bool) {
	parts := strings.Split(wk.Name, ": ")
	if len(parts) != 2 {
		return "", false
	}

	return parts[1], true
}

func (m *Manager) ProjectWks(project, display string) (i3.Workspace, bool, error) {
	wks, err := i3.Workspaces()
	if err != nil {
		return i3.Workspace{}, false, err
	}

	for _, w := range wks {
		wkProject, ok := workspaceProject(w)
		if !ok {
			continue
		}

		if wkProject != project {
			continue
		}

		if w.Output != display {
			continue
		}

		return w, true, nil
	}

	return i3.Workspace{}, false, nil
}

func (m *Manager) CurrentProject() (string, bool, error) {
	wks, err := i3.Workspaces()
	if err != nil {
		return "", false, err
	}

	for _, w := range wks {
		if !w.Visible {
			continue
		}

		project, ok := workspaceProject(w)
		if ok {
			return project, true, nil
		}
	}

	return "", false, nil
}

func (m *Manager) IsProjectVisble(project string) (bool, error) {
	i := 0
	wks, err := i3.Workspaces()
	if err != nil {
		return false, err
	}
	for _, w := range wks {
		if !w.Visible {
			continue
		}

		p, ok := workspaceProject(w)
		if ok && p == project {
			i++
		}
	}

	return i == 2, nil
}

func (m *Manager) OpenProjects() ([]string, error) {
	pmap := map[string]bool{}

	wks, err := i3.Workspaces()
	if err != nil {
		return nil, err
	}

	for _, w := range wks {
		if !i3.WorkspaceHasWindows(w.Name) {
			continue
		}

		project, ok := workspaceProject(w)
		if ok {
			pmap[project] = true
		}
	}

	projects := []string{}
	for k := range pmap {
		projects = append(projects, k)
	}

	sort.Strings(projects)

	return projects, nil
}

var minWorkspace = int64(workspaceStart)

func (m *Manager) nextWorkspacesID() (int64, error) {
	wks, err := i3.Workspaces()
	if err != nil {
		return 0, err
	}

	n := minWorkspace
	for _, w := range wks {
		if !i3.WorkspaceHasWindows(w.Name) {
			continue
		}

		_, ok := workspaceProject(w)
		if !ok {
			continue
		}

		if w.Num > n {
			n = w.Num
		}
	}

	n++

	minWorkspace = n

	return n, nil
}

func (m *Manager) displayActiveWorkspace(display string) (string, error) {
	allWks, err := i3.Workspaces()
	if err != nil {
		return "", err
	}
	for _, w := range allWks {
		if w.Visible && w.Output == display {
			return w.Name, nil
		}
	}

	return "", nil
}
