package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

const winSpawnTimeout = 5 * time.Second

func (m *Manager) OpenProject(project string) error {
	log.Printf("opening project %s", project)

	if m.SetupCommand != "" {
		if _, err := runShell(m.SetupCommand, projectEnv(project)); err != nil {
			return fmt.Errorf("setting up project %s: %w", project, err)
		}
	}

	wks, err := m.managedWorkspaces()
	if err != nil {
		return err
	}
	tree, err := i3.Tree(m.sw)
	if err != nil {
		return err
	}

	for _, cfg := range m.Workspaces {
		wk, ok := findProjectWks(wks, project, cfg.Display)
		if !ok {
			name, err := m.createProjectWks(project, cfg)
			if err != nil {
				return err
			}
			if err := m.spawnWindow(project, cfg, name); err != nil {
				return err
			}
			continue
		}

		hasWindows := i3.TreeWorkspaceHasWindows(tree, wk.Name)
		if !wk.Visible || !hasWindows {
			if err := i3.SwitchToWorkspace(m.sw, wk.Name); err != nil {
				return err
			}
		}
		if !hasWindows {
			if err := m.spawnWindow(project, cfg, wk.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) createProjectWks(project string, cfg WorkspaceConfig) (string, error) {
	num, err := m.nextWorkspaceNum()
	if err != nil {
		return "", err
	}
	label, err := m.workspaceLabel(project)
	if err != nil {
		return "", err
	}
	name := formatWorkspaceName(num, label, wksInfo{Project: project, Display: cfg.Display})

	output, err := m.resolveOutput(cfg.Display)
	if err != nil {
		return "", err
	}
	if output != "" {
		if err := i3.AssignWorkspace(m.sw, name, output); err != nil {
			return "", err
		}
	}

	if err := i3.SwitchToWorkspace(m.sw, name); err != nil {
		return "", err
	}
	return name, nil
}

func (m *Manager) spawnWindow(project string, cfg WorkspaceConfig, wksName string) error {
	events, stop, err := i3.WinEvents()
	if err != nil {
		return err
	}
	defer stop()

	if err := spawn(cfg.Command, projectEnv(project)); err != nil {
		return err
	}

	timeout := time.After(winSpawnTimeout)
	for {
		select {
		case ev := <-events:
			if i3.IsNewTiledWindow(ev) {
				return i3.MoveContainerToWorkspace(m.sw, ev.Container.ID, wksName)
			}
		case <-timeout:
			return fmt.Errorf("command %q did not spawn any new window after %s", cfg.Command, winSpawnTimeout)
		}
	}
}

func (m *Manager) workspaceLabel(project string) (string, error) {
	if m.WksNameCommand == "" {
		return project, nil
	}

	out, err := runShell(m.WksNameCommand, projectEnv(project))
	if err != nil {
		return "", fmt.Errorf("getting workspace name: %w", err)
	}
	if label := strings.TrimSpace(out); label != "" {
		return label, nil
	}
	return project, nil
}

func (m *Manager) ProjectList() ([]string, error) {
	if m.ListCommand == "" {
		return nil, errors.New("-list-cmd is required")
	}
	return listProjects(m.ListCommand)
}

func listProjects(command string) ([]string, error) {
	out, err := runShell(command, nil)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	return parseLines(out), nil
}

func parseLines(out string) []string {
	var res []string
	for line := range strings.Lines(out) {
		if p := strings.TrimSpace(line); p != "" {
			res = append(res, p)
		}
	}
	slices.Sort(res)
	return slices.Compact(res)
}

func projectEnv(project string) []string {
	return append(os.Environ(), "PROJECT_NAME="+project)
}

func spawn(command string, env []string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %q: %w", command, err)
	}
	return cmd.Process.Release()
}

func runShell(command string, env []string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}
	return stdout.String(), nil
}
