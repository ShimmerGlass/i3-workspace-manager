package main

import (
	"bytes"
	"encoding/gob"
	"sort"
	"strings"

	"github.com/eknkc/basex"
	"github.com/shimmerglass/i3-workspace-manager/i3"
)

const (
	workspaceStart = 20
)

var encoding *basex.Encoding

func init() {
	e, err := basex.NewEncoding("\uFEFF\u200D")
	if err != nil {
		panic(err)
	}
	encoding = e
}

type wksInfo struct {
	Project string
	Display string
}

func encodeWorkspaceName(name string, info wksInfo) string {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(info)
	return name + "\u200B" + encoding.Encode(buf.Bytes())
}

func decodeWorkspaceName(wk i3.Workspace) (wksInfo, bool, error) {
	res := wksInfo{}

	parts := strings.Split(wk.Name, "\u200B")
	if len(parts) != 2 {
		return res, false, nil
	}

	bin, err := encoding.Decode(parts[1])
	if err != nil {
		return res, false, err
	}

	buf := bytes.NewBuffer(bin)
	err = gob.NewDecoder(buf).Decode(&res)
	if err != nil {
		return res, false, err
	}

	return res, true, nil
}

func (m *Manager) ProjectWks(project, display string) (i3.Workspace, bool, error) {
	wks, err := i3.Workspaces()
	if err != nil {
		return i3.Workspace{}, false, err
	}

	for _, w := range wks {
		wkInfo, ok, err := decodeWorkspaceName(w)
		if err != nil {
			return i3.Workspace{}, false, err
		}
		if !ok {
			continue
		}

		if wkInfo.Project != project {
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

		wkInfo, ok, err := decodeWorkspaceName(w)
		if err != nil {
			return "", false, err
		}
		if ok {
			return wkInfo.Project, true, nil
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

		wkInfo, ok, err := decodeWorkspaceName(w)
		if err != nil {
			return false, err
		}

		if ok && wkInfo.Project == project {
			i++
		}
	}

	return i == len(m.Workspaces), nil
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

		wkInfo, ok, err := decodeWorkspaceName(w)
		if err != nil {
			return nil, err
		}
		if ok {
			pmap[wkInfo.Project] = true
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

	usedNums := map[int64]bool{}
	for _, w := range wks {
		usedNums[w.Num] = true
	}

	for n := minWorkspace; ; n++ {
		if usedNums[n] {
			continue
		}
		minWorkspace = n + 1
		return n, nil
	}
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
