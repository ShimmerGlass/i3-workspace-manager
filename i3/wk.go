package i3

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/joshuarubin/go-sway"
)

type (
	Client    = sway.Client
	Workspace = sway.Workspace
	Output    = sway.Output
	Node      = sway.Node
)

var quoteReplacer = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func Quote(s string) string {
	return `"` + quoteReplacer.Replace(s) + `"`
}

func Workspaces(sw Client) ([]Workspace, error) {
	return sw.GetWorkspaces(context.Background())
}

func Outputs(sw Client) ([]Output, error) {
	return sw.GetOutputs(context.Background())
}

func Tree(sw Client) (*Node, error) {
	return sw.GetTree(context.Background())
}

func run(sw Client, format string, args ...any) error {
	cmd := fmt.Sprintf(format, args...)
	log.Printf("sway: %s", cmd)
	_, err := sw.RunCommand(context.Background(), cmd)
	return err
}

func SwitchToWorkspace(sw Client, name string) error {
	return run(sw, "workspace %s", Quote(name))
}

func AssignWorkspace(sw Client, name, output string) error {
	return run(sw, "workspace %s output %s", Quote(name), Quote(output))
}

func CloseWorkspace(sw Client, name string) error {
	tree, err := Tree(sw)
	if err != nil {
		return err
	}
	for _, v := range TreeWorkspaceViews(tree, name) {
		if err := run(sw, "[con_id=%d] kill", v.ID); err != nil {
			return err
		}
	}
	return nil
}

func MoveCurrentWorkspace(sw Client, output string) error {
	return run(sw, "move workspace to output %s", Quote(output))
}

func MoveContainerToWorkspace(sw Client, conID int64, name string) error {
	return run(sw, "[con_id=%d] move container to workspace %s", conID, Quote(name))
}

func WorkspaceHasWindows(sw Client, name string) (bool, error) {
	tree, err := Tree(sw)
	if err != nil {
		return false, err
	}
	return TreeWorkspaceHasWindows(tree, name), nil
}

func TreeWorkspaceHasWindows(tree *Node, name string) bool {
	return len(TreeWorkspaceViews(tree, name)) > 0
}

func TreeWorkspaceViews(tree *Node, name string) []*Node {
	wk := tree.TraverseNodes(func(n *Node) bool {
		return n.Type == sway.NodeWorkspace && n.Name == name
	})
	if wk == nil {
		return nil
	}
	var views []*Node
	var walk func(n *Node)
	walk = func(n *Node) {
		if n.Sticky {
			return
		}
		if isView(n) {
			views = append(views, n)
			return
		}
		for _, c := range slices.Concat(n.Nodes, n.FloatingNodes) {
			walk(c)
		}
	}
	for _, c := range slices.Concat(wk.Nodes, wk.FloatingNodes) {
		walk(c)
	}
	return views
}

func isView(n *Node) bool {
	return (n.Type == sway.NodeCon || n.Type == sway.NodeFloatingCon) && n.PID != nil
}
