package i3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	i3 "go.i3wm.org/i3/v4"
)

type Workspace struct {
	Num     int    `json:"num,omitempty"`
	Name    string `json:"name,omitempty"`
	Output  string `json:"output,omitempty"`
	Visible bool   `json:"visible,omitempty"`
	Focused bool   `json:"focused,omitempty"`
}

func Workspaces() ([]Workspace, error) {
	out := &bytes.Buffer{}
	cmd := exec.Command("i3-msg", "-t", "get_workspaces")
	cmd.Stdout = out
	cmd.Stderr = out

	err := cmd.Run()
	if err != nil && out.Len() > 0 {
		return nil, fmt.Errorf("error getting workspaces: %s: %s", err, out.String())
	} else if err != nil {
		return nil, fmt.Errorf("error getting workspaces: %s", err)
	}

	res := []Workspace{}
	err = json.NewDecoder(out).Decode(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func WorkspaceByNum(i int) (*Workspace, error) {
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
	return Exec(fmt.Sprintf("workspace %s", name))
}

func CloseWorkspace(num int) error {
	log.Printf("closing to workspace %d", num)
	return Exec(fmt.Sprintf("[workspace=^%d] kill", num))
}

func RenameWorkspace(num int, title string) error {
	log.Printf("renaming workspace %d to %s", num, title)
	return Exec(
		fmt.Sprintf("rename workspace %d to \"%s\"", num, title),
	)
}

func WorkspaceHasWindows(name string) bool {
	tree, _ := i3.GetTree()
	wk := tree.Root.FindChild(func(n *i3.Node) bool {
		return n.Type == i3.WorkspaceNode && n.Name == name
	})

	return wk.FindChild(func(n *i3.Node) bool {
		return n.Type == i3.Con && n.Name != ""
	}) != nil

}
