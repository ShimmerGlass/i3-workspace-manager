package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/shimmerglass/i3-workspace-manager/i3"
	"github.com/shimmerglass/i3-workspace-manager/sel"
)

func (m *Manager) OpenProject(project string) error {
	wks, err := m.ProjectWks(project)
	if err != nil {
		return err
	}

	err = m.setupProject(project)
	if err != nil {
		return err
	}

	for i, w := range wks {
		if w != nil {
			i3.SwitchToWorkspace(w.Name)
		} else {
			wkn, err := m.nextWorkspacesID(i)
			if err != nil {
				return err
			}

			err = i3.SwitchToWorkspace(strconv.Itoa(wkn))
			if err != nil {
				return err
			}
			err = i3.RenameWorkspace(wkn, fmt.Sprintf("%d: %s", wkn, project))
			if err != nil {
				return err
			}

			cmd := exec.Command("/bin/sh", "-c", m.Workspaces[i].Command)
			env := os.Environ()
			stderr := &bytes.Buffer{}
			cmd.Stderr = stderr
			env = append(env, fmt.Sprintf("PROJECT_NAME=%s", project))
			cmd.Env = env
			err = cmd.Run()
			if err != nil {
				if len(stderr.Bytes()) > 0 {
					return fmt.Errorf(string(stderr.Bytes()))
				}
				return err
			}

			err = i3.Exec(fmt.Sprintf("move workspace to output %s", m.Workspaces[i].Display))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) setupProject(name string) error {
	if m.SetupCommand == "" {
		return nil
	}

	cmd := exec.Command("/bin/sh", "-c", m.SetupCommand)
	env := os.Environ()
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	env = append(env, fmt.Sprintf("PROJECT_NAME=%s", name))
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		if len(stderr.Bytes()) > 0 {
			return fmt.Errorf(string(stderr.Bytes()))
		}
		return err
	}
	return nil
}

func (m *Manager) projectList() (string, error) {
	cmd := exec.Command("/bin/sh", "-c", m.ListCommand)
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		if len(stderr.Bytes()) > 0 {
			return "", fmt.Errorf(string(stderr.Bytes()))
		}
		return "", err
	}

	fmt.Println(string(stdout.Bytes()))

	return sel.Do(stdout)
}
