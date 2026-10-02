package main

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"slices"
	"strings"

	"github.com/eknkc/basex"
	"github.com/shimmerglass/i3-workspace-manager/i3"
)

const (
	workspaceStart = 20
	metaSeparator  = "\u200B"
)

var encoding = mustEncoding("\uFEFF\u200C\u200D\u2060")

func mustEncoding(alphabet string) *basex.Encoding {
	e, err := basex.NewEncoding(alphabet)
	if err != nil {
		panic(err)
	}
	return e
}

type wksInfo struct {
	Project string
	Display string
}

type managedWks struct {
	i3.Workspace
	wksInfo
}

func encodeWorkspaceName(label string, info wksInfo) string {
	buf := &bytes.Buffer{}
	writeLenString(buf, info.Project)
	writeLenString(buf, info.Display)
	return label + metaSeparator + encoding.Encode(buf.Bytes())
}

func parseWorkspaceName(name string) (wksInfo, bool, error) {
	_, meta, ok := strings.Cut(name, metaSeparator)
	if !ok || strings.Contains(meta, metaSeparator) {
		return wksInfo{}, false, nil
	}

	bin, err := encoding.Decode(meta)
	if err != nil {
		return wksInfo{}, false, err
	}

	buf := bytes.NewBuffer(bin)
	var res wksInfo
	if res.Project, err = readLenString(buf); err != nil {
		return wksInfo{}, false, err
	}
	if res.Display, err = readLenString(buf); err != nil {
		return wksInfo{}, false, err
	}
	if buf.Len() != 0 {
		return wksInfo{}, false, errors.New("trailing data")
	}

	return res, true, nil
}

func decodeWorkspaceName(name string) (wksInfo, bool) {
	info, ok, err := parseWorkspaceName(name)
	if err != nil {
		log.Printf("ignoring workspace %q with invalid metadata: %s", name, err)
	}
	return info, ok
}

func writeLenString(buf *bytes.Buffer, s string) {
	buf.Write(binary.AppendUvarint(nil, uint64(len(s))))
	buf.WriteString(s)
}

func readLenString(buf *bytes.Buffer) (string, error) {
	n, err := binary.ReadUvarint(buf)
	if err != nil {
		return "", err
	}
	if n > uint64(buf.Len()) {
		return "", io.ErrUnexpectedEOF
	}
	return string(buf.Next(int(n))), nil
}

func managedWorkspaces(wks []i3.Workspace) []managedWks {
	var res []managedWks
	for _, w := range wks {
		if info, ok := decodeWorkspaceName(w.Name); ok {
			res = append(res, managedWks{Workspace: w, wksInfo: info})
		}
	}
	return res
}

func (m *Manager) managedWorkspaces() ([]managedWks, error) {
	wks, err := i3.Workspaces(m.sw)
	if err != nil {
		return nil, err
	}
	return managedWorkspaces(wks), nil
}

func findProjectWks(wks []managedWks, project, display string) (managedWks, bool) {
	for _, w := range wks {
		if w.Project == project && w.Display == display {
			return w, true
		}
	}
	return managedWks{}, false
}

func (m *Manager) CurrentProject() (string, bool, error) {
	wks, err := m.managedWorkspaces()
	if err != nil {
		return "", false, err
	}
	project, ok := currentProject(wks)
	return project, ok, nil
}

func currentProject(wks []managedWks) (string, bool) {
	for _, w := range wks {
		if w.Focused {
			return w.Project, true
		}
	}
	for _, w := range wks {
		if w.Visible {
			return w.Project, true
		}
	}
	return "", false
}

func (m *Manager) IsProjectVisible(project string) (bool, error) {
	wks, err := m.managedWorkspaces()
	if err != nil {
		return false, err
	}
	return projectVisible(wks, project, len(m.Workspaces)), nil
}

func projectVisible(wks []managedWks, project string, expected int) bool {
	count := 0
	shown := map[string]bool{}
	for _, w := range wks {
		if w.Project != project {
			continue
		}
		count++
		shown[w.Output] = shown[w.Output] || w.Visible
	}

	if count == 0 || count != expected {
		return false
	}
	for _, ok := range shown {
		if !ok {
			return false
		}
	}
	return true
}

func (m *Manager) OpenProjects() ([]string, error) {
	wks, err := m.managedWorkspaces()
	if err != nil {
		return nil, err
	}
	tree, err := i3.Tree(m.sw)
	if err != nil {
		return nil, err
	}

	open := map[string]bool{}
	for _, w := range wks {
		if i3.TreeWorkspaceHasWindows(tree, w.Name) {
			open[w.Project] = true
		}
	}
	return slices.Sorted(maps.Keys(open)), nil
}

func (m *Manager) resolveOutput(role string) (string, error) {
	outputs, err := i3.Outputs(m.sw)
	if err != nil {
		return "", err
	}
	return resolveRole(activeOutputsSorted(outputs), role), nil
}

func activeOutputsSorted(outputs []i3.Output) []i3.Output {
	active := slices.DeleteFunc(slices.Clone(outputs), func(o i3.Output) bool { return !o.Active })
	slices.SortFunc(active, func(a, b i3.Output) int { return cmp.Compare(a.Rect.X, b.Rect.X) })
	return active
}

func resolveRole(active []i3.Output, role string) string {
	for _, o := range active {
		if o.Name == role {
			return o.Name
		}
	}
	if len(active) == 0 {
		return ""
	}

	switch role {
	case "left":
		return active[0].Name
	case "right":
		return active[len(active)-1].Name
	default:
		return ""
	}
}

func (m *Manager) reconcileDisplays() error {
	wks, err := i3.Workspaces(m.sw)
	if err != nil {
		return err
	}
	outputs, err := i3.Outputs(m.sw)
	if err != nil {
		return err
	}
	active := activeOutputsSorted(outputs)

	var errs []error
	moved := false
	for _, w := range managedWorkspaces(wks) {
		if w.Display == "" {
			continue
		}

		target := resolveRole(active, w.Display)
		if target == "" || w.Output == target {
			continue
		}

		log.Printf("re-pinning workspace %s to %s (role %q, currently on %s)", w.Name, target, w.Display, w.Output)
		if err := i3.SwitchToWorkspace(m.sw, w.Name); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := i3.MoveCurrentWorkspace(m.sw, target); err != nil {
			errs = append(errs, err)
			continue
		}
		moved = true
	}

	if moved {
		errs = append(errs, m.restoreVisible(wks))
	}
	return errors.Join(errs...)
}

func (m *Manager) restoreVisible(before []i3.Workspace) error {
	current, err := i3.Workspaces(m.sw)
	if err != nil {
		return err
	}
	exists := map[string]bool{}
	for _, w := range current {
		exists[w.Name] = true
	}

	var errs []error
	focused := ""
	for _, w := range before {
		switch {
		case w.Focused:
			focused = w.Name
		case w.Visible && exists[w.Name]:
			errs = append(errs, i3.SwitchToWorkspace(m.sw, w.Name))
		}
	}
	if focused != "" {
		errs = append(errs, i3.SwitchToWorkspace(m.sw, focused))
	}
	return errors.Join(errs...)
}

func (m *Manager) nextWorkspaceNum() (int64, error) {
	wks, err := i3.Workspaces(m.sw)
	if err != nil {
		return 0, err
	}

	used := map[int64]bool{}
	for _, w := range wks {
		used[w.Num] = true
	}

	n := max(m.nextWksNum, workspaceStart)
	for used[n] {
		n++
	}
	m.nextWksNum = n + 1
	return n, nil
}

func formatWorkspaceName(num int64, label string, info wksInfo) string {
	return fmt.Sprintf("%d:%s", num, encodeWorkspaceName(label, info))
}
