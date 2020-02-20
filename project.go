package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (m *Manager) ensureProject(path, name string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	err = cmd("git", "clone", "ssh://review.criteois.lan:29418/"+name, path)
	if err != nil {
		return err
	}
	err = cmd("scp", "-p", "-P", "29418", "review.criteois.lan:hooks/commit-msg", filepath.Join(path, ".git/hooks/"))
	if err != nil {
		return err
	}

	return nil
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
