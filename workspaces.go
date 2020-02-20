package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

type i3Workspace struct {
	Visible bool   `json:"visible,omitempty"`
	Name    string `json:"name,omitempty"`
}

type workspace struct {
	Num    int    `json:"num,omitempty"`
	Name   string `json:"name,omitempty"`
	Output string `json:"output,omitempty"`
}

func workspaces() ([]workspace, error) {
	out := &bytes.Buffer{}
	cmd := exec.Command("i3-msg", "-t", "get_workspaces")
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	res := []workspace{}
	err = json.NewDecoder(out).Decode(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Manager) workspaceForProject(project string) (workspace, workspace, bool, error) {
	wks, err := workspaces()
	if err != nil {
		return workspace{}, workspace{}, false, err
	}

	var left, right workspace
	var leftOk, rightOk bool
	for _, w := range wks {
		if strings.HasSuffix(w.Name, project) {
			switch w.Output {
			case m.leftDisplay:
				left = w
				leftOk = true
			case m.rightDisplay:
				right = w
				rightOk = true
			}
		}
	}

	return left, right, leftOk && rightOk, nil
}

func (m *Manager) nextWorkspaces(leftDisplay, rightDisplay string) (int, int, error) {
	wks, err := workspaces()
	if err != nil {
		return 0, 0, err
	}

	left, right := 20, 40
	for _, w := range wks {
		if strings.Contains(w.Name, ":") {
			switch w.Output {
			case leftDisplay:
				left = w.Num + 1
			case rightDisplay:
				right = w.Num + 1
			}
		}
	}

	return left, right, nil
}

func (m *Manager) currentProject() (string, error) {
	cmd := exec.Command("i3-msg", "-t", "get_workspaces")
	err := cmd.Start()
	if err != nil {
		return "", err
	}

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	wks := []i3Workspace{}
	err = json.Unmarshal(out, &wks)
	if err != nil {
		return "", err
	}

	for _, wk := range wks {
		if !wk.Visible {
			continue
		}

		parts := strings.Split(wk.Name, ": ")
		if len(parts) != 2 {
			continue
		}

		return parts[1], nil
	}

	return "", nil
}
