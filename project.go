package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/shimmerglass/i3-workspace-manager/i3"
	"github.com/shimmerglass/i3-workspace-manager/sel"
)

func (m *Manager) OpenProject(project string) error {
	log.Printf("opening project %s", project)

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
			if !w.Visible {
				err := i3.SwitchToWorkspace(w.Name)
				if err != nil {
					return err
				}
			}
		} else {
			wkn, err := m.nextWorkspacesID()
			if err != nil {
				return err
			}

			displayActiveWorkspace, err := m.displayActiveWorkspace(m.Workspaces[i].Display)
			if err != nil {
				return err
			}

			if displayActiveWorkspace != "" {
				err = i3.SwitchToWorkspace(displayActiveWorkspace)
				if err != nil {
					return err
				}

				time.Sleep(100 * time.Millisecond)
			}

			name := fmt.Sprintf("%d: %s", wkn, project)

			err = i3.SwitchToWorkspace(name)
			if err != nil {
				return err
			}

			time.Sleep(100 * time.Millisecond)

			if displayActiveWorkspace == "" {
				err = i3.Exec(fmt.Sprintf("move workspace to output %s", m.Workspaces[i].Display))
				if err != nil {
					return err
				}

				time.Sleep(100 * time.Millisecond)
			}

			cmd := exec.Command("/bin/sh", "-c", m.Workspaces[i].Command)
			env := os.Environ()
			out := &bytes.Buffer{}
			cmd.Stderr = out
			cmd.Stdout = out
			env = append(env, fmt.Sprintf("PROJECT_NAME=%s", project))
			cmd.Env = env
			err = cmd.Run()
			if err != nil {
				if out.Len() > 0 {
					return fmt.Errorf("error opening workspace with command %s: %s", m.Workspaces[i].Command, out.String())
				}
				return fmt.Errorf("error setting up workspace: %s", err)
			}

			time.Sleep(100 * time.Millisecond)

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
		return fmt.Errorf("error setting up project: %s", err)
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
		return "", fmt.Errorf("error listing projects: %s", err)
	}

	return sel.Do(stdout)
}
