package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
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
	// The metadata is stored in the workspace name using zero-width code points
	// so it stays invisible in bars and window titles. A larger alphabet means
	// fewer characters per byte: base 4 emits 4 chars/byte instead of the 8 a
	// base-2 alphabet would. All four are reliably zero-width and distinct from
	// the "\u200B" separator used below.
	e, err := basex.NewEncoding("\uFEFF\u200C\u200D\u2060")
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
	writeLenString(buf, info.Project)
	writeLenString(buf, info.Display)
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
	if res.Project, err = readLenString(buf); err != nil {
		return res, false, err
	}
	if res.Display, err = readLenString(buf); err != nil {
		return res, false, err
	}

	return res, true, nil
}

// writeLenString writes s as a varint length prefix followed by its raw bytes.
// This is a compact, self-describing alternative to gob, which would emit its
// type descriptor on every call and dwarf the few bytes of actual payload.
func writeLenString(buf *bytes.Buffer, s string) {
	var l [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(l[:], uint64(len(s)))
	buf.Write(l[:n])
	buf.WriteString(s)
}

func readLenString(buf *bytes.Buffer) (string, error) {
	n, err := binary.ReadUvarint(buf)
	if err != nil {
		return "", err
	}
	s := make([]byte, n)
	if _, err := io.ReadFull(buf, s); err != nil {
		return "", err
	}
	return string(s), nil
}

func (m *Manager) ProjectWks(project, display string) (i3.Workspace, bool, error) {
	wks, err := i3.Workspaces(m.sw)
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

		if wkInfo.Display != display {
			continue
		}

		return w, true, nil
	}

	return i3.Workspace{}, false, nil
}

func (m *Manager) CurrentProject() (string, bool, error) {
	wks, err := i3.Workspaces(m.sw)
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
	wks, err := i3.Workspaces(m.sw)
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

	wks, err := i3.Workspaces(m.sw)
	if err != nil {
		return nil, err
	}

	for _, w := range wks {
		if !i3.WorkspaceHasWindows(m.sw, w.Name) {
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

// resolveOutput maps a display role stored in a workspace name to the name of a
// currently connected output. The roles "left" and "right" resolve to the
// leftmost and rightmost active output by position, so pinning follows the
// physical layout even when the two identical panels bind to their connectors
// in a different order at boot (or after sway-swap-monitors flips them). For
// backward compatibility a role that is itself the name of a connected output
// (e.g. "DP-2") is used verbatim. Returns "" when no output can satisfy the
// role (e.g. the target display is disconnected).
func (m *Manager) resolveOutput(role string) (string, error) {
	outputs, err := i3.Outputs(m.sw)
	if err != nil {
		return "", err
	}

	active := []i3.Output{}
	for _, o := range outputs {
		if o.Active {
			active = append(active, o)
		}
	}
	// Backward compatibility: role given as a literal output name.
	for _, o := range active {
		if o.Name == role {
			return o.Name, nil
		}
	}
	if len(active) == 0 {
		return "", nil
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].Rect.X < active[j].Rect.X
	})
	switch role {
	case "left":
		return active[0].Name, nil
	case "right":
		return active[len(active)-1].Name, nil
	default:
		return "", fmt.Errorf("unknown display role %q", role)
	}
}

// reconcileDisplays moves every project workspace back to the output that
// satisfies the display role encoded in its name. When a display disconnects,
// sway evacuates its workspaces to a surviving output and does not move them
// back on reconnect; this restores them. It also runs after output position
// changes (e.g. sway-swap-monitors), re-pinning role-based workspaces so they
// follow the panel that is now leftmost/rightmost. Workspaces whose role cannot
// currently be resolved are left where they are.
func (m *Manager) reconcileDisplays() error {
	wks, err := i3.Workspaces(m.sw)
	if err != nil {
		return err
	}

	focused := ""
	for _, w := range wks {
		if w.Focused {
			focused = w.Name
		}
	}

	moved := false
	for _, w := range wks {
		info, ok, err := decodeWorkspaceName(w)
		if err != nil {
			return err
		}
		if !ok || info.Display == "" {
			continue
		}

		target, err := m.resolveOutput(info.Display)
		if err != nil {
			log.Printf("cannot resolve display role %q for workspace %s: %s", info.Display, w.Name, err)
			continue
		}
		// Role unsatisfiable (display disconnected), or already there.
		if target == "" || w.Output == target {
			continue
		}

		log.Printf("re-pinning workspace %s to %s (role %q, currently on %s)", w.Name, target, info.Display, w.Output)
		// A workspace can only be moved to an output while focused, so switch to
		// it first, then move the now-focused workspace.
		if err := i3.SwitchToWorkspace(m.sw, w.Name); err != nil {
			return err
		}
		if err := i3.MoveCurrentWorkspace(m.sw, target); err != nil {
			return err
		}
		moved = true
	}

	// Restore the workspace that was focused before we started shuffling.
	if moved && focused != "" {
		return i3.SwitchToWorkspace(m.sw, focused)
	}

	return nil
}

var minWorkspace = int64(workspaceStart)

func (m *Manager) nextWorkspacesID() (int64, error) {
	wks, err := i3.Workspaces(m.sw)
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
