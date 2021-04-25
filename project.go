package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

const winSpawnTimeout = 5 * time.Second

func (m *Manager) OpenProject(project string) error {
	log.Printf("opening project %s", project)

	err := m.setupProject(project)
	if err != nil {
		return err
	}

	for _, cfg := range m.Workspaces {
		wk, ok, err := m.ProjectWks(project, cfg.Display)
		if err != nil {
			return err
		}

		if ok {
			hasWindows := i3.WorkspaceHasWindows(wk.Name)
			if !wk.Visible || !hasWindows {
				err := i3.SwitchToWorkspace(wk.Name)
				if err != nil {
					return err
				}
			}
			if !hasWindows {
				err = m.openProjectDisplay(project, cfg)
				if err != nil {
					return err
				}
			}
		} else {
			err := m.openProjectCreateWks(project, cfg)
			if err != nil {
				return err
			}
			err = m.openProjectDisplay(project, cfg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) openProjectCreateWks(project string, cfg Workspace) error {
	wkn, err := m.nextWorkspacesID()
	if err != nil {
		return err
	}

	activeWks, err := m.displayActiveWorkspace(cfg.Display)
	if err != nil {
		return err
	}

	if activeWks != "" {
		err = i3.SwitchToWorkspace(activeWks)
		if err != nil {
			return err
		}
	}

	wksName := fmt.Sprintf("%d: %s", wkn, project)
	err = i3.SwitchToWorkspace(wksName)
	if err != nil {
		return err
	}

	if activeWks == "" {
		err = i3.MoveCurrentWorkspace(cfg.Display)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) openProjectDisplay(project string, cfg Workspace) error {
	winEvts, done := i3.WinEvents()
	defer done()

	err := m.runOpenCommand(project, cfg.Command)
	if err != nil {
		return err
	}

	timeout := time.NewTimer(winSpawnTimeout)
WaitWindow:
	for {
		select {
		case ev := <-winEvts:
			log.Printf("i3 win event: %+v", ev)
			if ev.Change == "new" {
				break WaitWindow
			}
		case <-timeout.C:
			return fmt.Errorf("command %q did not spawn any new window after %s", cfg.Command, winSpawnTimeout)
		}
	}

	return nil
}

func (m *Manager) runOpenCommand(project, command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	env := os.Environ()
	out := &bytes.Buffer{}
	cmd.Stderr = out
	cmd.Stdout = out
	env = append(env, fmt.Sprintf("PROJECT_NAME=%s", project))
	cmd.Env = env
	err := cmd.Start()
	if err != nil {
		if out.Len() > 0 {
			return fmt.Errorf("error opening workspace with command %s: %s", command, out.String())
		}
		return fmt.Errorf("error setting up workspace: %s", err)
	}

	return cmd.Process.Release()
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

func (m *Manager) ProjectList() ([]string, error) {
	cmd := exec.Command("/bin/sh", "-c", m.ListCommand)
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		if len(stderr.Bytes()) > 0 {
			return nil, fmt.Errorf(string(stderr.Bytes()))
		}
		return nil, fmt.Errorf("error listing projects: %s", err)
	}

	res := []string{}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		p := scanner.Text()
		if p == "" {
			continue
		}
		res = append(res, strings.TrimSpace(p))
	}

	sort.Strings(res)

	return res, nil
}
