package i3

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
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

func RunCommand(sw Client, cmd string) error {
	return run(sw, "%s", cmd)
}

func FindContainer(tree *Node, conID int64) *Node {
	if tree.ID == conID {
		return tree
	}
	for _, c := range slices.Concat(tree.Nodes, tree.FloatingNodes) {
		if n := FindContainer(c, conID); n != nil {
			return n
		}
	}
	return nil
}

func ContainerWorkspace(tree *Node, conID int64) (string, bool) {
	var found string
	var ok bool
	var walk func(n *Node, wks string)
	walk = func(n *Node, wks string) {
		if n.Type == sway.NodeWorkspace {
			wks = n.Name
		}
		if n.ID == conID && n.PID != nil {
			found, ok = wks, true
			return
		}
		for _, c := range slices.Concat(n.Nodes, n.FloatingNodes) {
			if ok {
				return
			}
			walk(c, wks)
		}
	}
	walk(tree, "")
	return found, ok
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
	views := TreeWorkspaceViews(tree, name)
	var scratch map[int64]bool
	if slices.ContainsFunc(views, func(v *Node) bool { return v.Type == sway.NodeFloatingCon }) {
		if scratch, err = scratchpadIDs(); err != nil {
			return err
		}
	}
	for _, v := range views {
		action := "kill"
		if v.Sticky || scratch[v.ID] {
			action = "move scratchpad"
		}
		if err := run(sw, "[con_id=%d] %s", v.ID, action); err != nil {
			return err
		}
	}
	return nil
}

type rawNode struct {
	ID              int64     `json:"id"`
	ScratchpadState string    `json:"scratchpad_state"`
	Nodes           []rawNode `json:"nodes"`
	FloatingNodes   []rawNode `json:"floating_nodes"`
}

func scratchpadIDs() (map[int64]bool, error) {
	out, err := exec.Command("swaymsg", "-r", "-t", "get_tree").Output()
	if err != nil {
		return nil, fmt.Errorf("swaymsg get_tree: %w", err)
	}
	return parseScratchpadIDs(out)
}

func parseScratchpadIDs(data []byte) (map[int64]bool, error) {
	var root rawNode
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	ids := map[int64]bool{}
	var walk func(n rawNode)
	walk = func(n rawNode) {
		if n.ScratchpadState != "" && n.ScratchpadState != "none" {
			ids[n.ID] = true
		}
		for _, c := range slices.Concat(n.Nodes, n.FloatingNodes) {
			walk(c)
		}
	}
	walk(root)
	return ids, nil
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
	return slices.ContainsFunc(TreeWorkspaceViews(tree, name), func(v *Node) bool { return !v.Sticky })
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
