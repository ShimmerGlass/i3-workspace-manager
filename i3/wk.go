package i3

import (
	"context"
	"fmt"
	"log"

	i3 "github.com/joshuarubin/go-sway"
)

type Workspace = i3.Workspace
type Output = i3.Output

func Workspaces(sw i3.Client) ([]Workspace, error) {
	return sw.GetWorkspaces(context.Background())
}

func Outputs(sw i3.Client) ([]Output, error) {
	return sw.GetOutputs(context.Background())
}

func WorkspaceByNum(sw i3.Client, i int64) (*Workspace, error) {
	wks, err := Workspaces(sw)
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

func WorkspaceByName(sw i3.Client, n string) (*Workspace, error) {
	wks, err := Workspaces(sw)
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

func SwitchToWorkspace(sw i3.Client, name string) error {
	log.Printf("switching to workspace %s", name)
	_, err := sw.RunCommand(context.Background(), fmt.Sprintf("workspace %s", name))
	return err
}

func AssignWorkspace(sw i3.Client, name, output string) error {
	log.Printf("switching to workspace %s", name)
	_, err := sw.RunCommand(context.Background(), fmt.Sprintf("workspace %s output %s", name, output))
	return err
}

func CloseWorkspace(sw i3.Client, num int64) error {
	log.Printf("closing workspace %d", num)
	_, err := sw.RunCommand(context.Background(), fmt.Sprintf(`[workspace="^%d" tiling] kill`, num))
	return err
}

func RenameWorkspace(sw i3.Client, num int, title string) error {
	log.Printf("renaming workspace %d to %s", num, title)
	_, err := sw.RunCommand(
		context.Background(),
		fmt.Sprintf("rename workspace %d to \"%s\"", num, title),
	)
	return err
}

func MoveCurrentWorkspace(sw i3.Client, display string) error {
	_, err := sw.RunCommand(context.Background(), fmt.Sprintf("move workspace to output %s", display))
	return err
}

func MoveContainerToWorkspace(sw i3.Client, conID int64, wksName string) error {
	log.Printf("moving container %d to workspace %s", conID, wksName)
	_, err := sw.RunCommand(
		context.Background(),
		fmt.Sprintf(`[con_id=%d] move container to workspace "%s"`, conID, wksName),
	)
	return err
}

func WorkspaceHasWindows(sw i3.Client, name string) bool {
	tree, _ := sw.GetTree(context.Background())
	wk := tree.TraverseNodes(func(n *i3.Node) bool {
		return n.Type == i3.NodeWorkspace && n.Name == name
	})
	if wk == nil {
		return false
	}
	for _, c := range wk.Nodes {
		return c.TraverseNodes(func(n *i3.Node) bool {
			return n.Type == i3.NodeCon && n.Name != ""
		}) != nil
	}
	return false
}
