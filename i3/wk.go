package i3

import (
	"fmt"
	"log"

	i3 "go.i3wm.org/i3/v4"
)

type Workspace = i3.Workspace

func Workspaces() ([]Workspace, error) {
	return i3.GetWorkspaces()
}

func WorkspaceByNum(i int64) (*Workspace, error) {
	wks, err := Workspaces()
	if err != nil {
		return nil, err
	}

	for _, w := range wks {
		if w.Num == i {
			return &w, nil
		}
	}

	return nil, nil
}

func WorkspaceByName(n string) (*Workspace, error) {
	wks, err := Workspaces()
	if err != nil {
		return nil, err
	}

	for _, w := range wks {
		if w.Name == n {
			return &w, nil
		}
	}

	return nil, nil
}

func SwitchToWorkspace(name string) error {
	log.Printf("switching to workspace %s", name)
	_, err := i3.RunCommand(fmt.Sprintf("workspace %s", name))
	return err
}

func CloseWorkspace(num int64) error {
	log.Printf("closing workspace %d", num)
	_, err := i3.RunCommand(fmt.Sprintf(`[workspace="^%d" tiling] kill`, num))
	return err
}

func RenameWorkspace(num int, title string) error {
	log.Printf("renaming workspace %d to %s", num, title)
	_, err := i3.RunCommand(
		fmt.Sprintf("rename workspace %d to \"%s\"", num, title),
	)
	return err
}

func MoveCurrentWorkspace(display string) error {
	_, err := i3.RunCommand(fmt.Sprintf("move workspace to output %s", display))
	return err
}

func WorkspaceHasWindows(name string) bool {
	tree, _ := i3.GetTree()
	wk := tree.Root.FindChild(func(n *i3.Node) bool {
		return n.Type == i3.WorkspaceNode && n.Name == name
	})
	if wk == nil {
		return false
	}
	for _, c := range wk.Nodes {
		return c.FindChild(func(n *i3.Node) bool {
			return n.Type == i3.Con && n.Name != ""
		}) != nil
	}
	return false
}
