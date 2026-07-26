package main

import (
	"bufio"
	"bytes"
	"errors"
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
			hasWindows := i3.WorkspaceHasWindows(m.sw, wk.Name)
			if !wk.Visible || !hasWindows {
				err := i3.SwitchToWorkspace(m.sw, wk.Name)
				if err != nil {
					return err
				}
			}
			if !hasWindows {
				err = m.openProjectDisplay(project, cfg, wk.Name)
				if err != nil {
					return err
				}
			}
		} else {
			wksName, err := m.openProjectCreateWks(project, cfg)
			if err != nil {
				return err
			}

			err = m.openProjectDisplay(project, cfg, wksName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) openProjectCreateWks(project string, cfg Workspace) (string, error) {
	wkn, err := m.nextWorkspacesID()
	if err != nil {
		return "", err
	}

	userWksName, err := m.workspaceName(project)
	if err != nil {
		return "", err
	}
	wksName := encodeWorkspaceName(userWksName, wksInfo{
		Project: project,
		Display: cfg.Display,
	})
	wksFullName := fmt.Sprintf("%d:%s", wkn, wksName)

	// Assign the workspace to its output *before* creating it. In sway,
	// `workspace <name> output <output>` only sets a placement rule for future
	// creation; it does not move a workspace that already exists. Creating the
	// workspace first (via SwitchToWorkspace) would place it on the currently
	// focused output instead, which is racy: a previously opened app (e.g. VS
	// Code) can steal focus back to another display before the next workspace
	// is created, landing it on the wrong output. cfg.Display is a display role
	// (e.g. "left"), so resolve it to a concrete output for the placement rule;
	// the role itself is what gets encoded into the name for later reconciling.
	output, err := m.resolveOutput(cfg.Display)
	if err != nil {
		return "", err
	}
	err = i3.AssignWorkspace(m.sw, wksFullName, output)
	if err != nil {
		return "", err
	}

	err = i3.SwitchToWorkspace(m.sw, wksFullName)
	if err != nil {
		return "", err
	}

	return wksFullName, nil
}

func (m *Manager) openProjectDisplay(project string, cfg Workspace, wksName string) error {
	winEvts, done, err := i3.WinEvents()
	if err != nil {
		return err
	}
	defer done()

	err = m.runOpenCommand(project, cfg.Command)
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
				// The window maps on whichever workspace currently has focus,
				// which is racy: a slow app opened for a previous display (e.g.
				// VS Code) can steal focus back before this window appears,
				// landing it on the wrong workspace. Move the newly spawned
				// container to its target workspace explicitly, by container id,
				// so placement does not depend on focus.
				err := i3.MoveContainerToWorkspace(m.sw, ev.Container.ID, wksName)
				if err != nil {
					return err
				}
				break WaitWindow
			}
		case <-timeout.C:
			return fmt.Errorf("command %q did not spawn any new window after %s", cfg.Command, winSpawnTimeout)
		}
	}

	return nil
}

func (m *Manager) runOpenCommand(project, command string) error {
	fmt.Println(command)
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
		if stderr.Len() > 0 {
			return errors.New(stderr.String())
		}
		return fmt.Errorf("error setting up project: %s", err)
	}
	return nil
}

func (m *Manager) workspaceName(name string) (string, error) {
	if m.WksNameCommand == "" {
		return name, nil
	}

	cmd := exec.Command("/bin/sh", "-c", m.WksNameCommand)
	env := os.Environ()
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	env = append(env, fmt.Sprintf("PROJECT_NAME=%s", name))
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(stderr.String())
		}
		return "", fmt.Errorf("error getting workspace name: %s", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (m *Manager) ProjectList() ([]string, error) {
	cmd := exec.Command("/bin/sh", "-c", m.ListCommand)
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, errors.New(stderr.String())
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
