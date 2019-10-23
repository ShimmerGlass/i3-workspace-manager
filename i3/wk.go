package i3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	res := []Workspace{}
	err = json.NewDecoder(out).Decode(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func SwitchToWorkspace(name string) error {
	return Exec(fmt.Sprintf("workspace %s", name))
}

func CloseWorkspace(num int) error {
	return Exec(fmt.Sprintf("[workspace=^%d:] kill", num))
}

func RenameWorkspace(num int, title string) error {
	return Exec(
		fmt.Sprintf("rename workspace %d to \"%s\"", num, title),
	)
}
