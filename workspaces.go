package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

const (
	leftDisplay  = "DP-5"
	rightDisplay = "DP-4"
)

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

func workspaceForProject(project string) (workspace, workspace, bool, error) {
	wks, err := workspaces()
	if err != nil {
		return workspace{}, workspace{}, false, err
	}

	var left, right workspace
	var leftOk, rightOk bool
	for _, w := range wks {
		if strings.HasSuffix(w.Name, project) {
			switch w.Output {
			case leftDisplay:
				left = w
				leftOk = true
			case rightDisplay:
				right = w
				rightOk = true
			}
		}
	}

	return left, right, leftOk && rightOk, nil
}

func nextWorkspaces() (int, int, error) {
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
