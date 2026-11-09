package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
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
			name, output, err := m.newProjectWks(project, cfg)
			if err != nil {
				return err
			}
			if err := m.spawnWindow(project, cfg, name, output, output); err != nil {
				return err
			}
			continue
		}

		if !i3.TreeWorkspaceHasWindows(tree, wk.Name) {
			if err := m.spawnWindow(project, cfg, wk.Name, "", wk.Output); err != nil {
				return err
			}
			continue
		}
		if !m.Background && !wk.Visible {
			if err := i3.SwitchToWorkspace(m.sw, wk.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) newProjectWks(project string, cfg WorkspaceConfig) (string, string, error) {
	num, err := m.nextWorkspaceNum()
	if err != nil {
		return "", "", err
	}
	label, err := m.workspaceLabel(project)
	if err != nil {
		return "", "", err
	}
	name := formatWorkspaceName(num, label, wksInfo{Project: project, Display: cfg.Display})

	output, err := m.resolveOutput(cfg.Display)
	if err != nil {
		return "", "", err
	}
	return name, output, nil
}

func (m *Manager) spawnWindow(project string, cfg WorkspaceConfig, wksName, assignOutput, wksOutput string) error {
	var restore []string
	if m.Background {
		current, err := i3.Workspaces(m.sw)
		if err != nil {
			return err
		}
		restore = restoreTargets(current, wksOutput)
	}

	command, err := buildSpawnCommand(wksName, assignOutput, project, cfg.Command, restore)
	if err != nil {
		return err
	}

	events, stop, err := i3.WinEvents()
	if err != nil {
		return err
	}
	defer stop()

	if err := i3.RunCommand(m.sw, command); err != nil {
		return err
	}

	var strays []int64
	timeout := time.After(winSpawnTimeout)
	for {
		select {
		case ev := <-events:
			if !i3.IsNewTiledWindow(ev) {
				continue
			}
			tree, err := i3.Tree(m.sw)
			if err != nil {
				return err
			}
			if name, ok := i3.ContainerWorkspace(tree, ev.Container.ID); ok && name == wksName {
				if !m.Background {
					return nil
				}
				return m.wakeIfStuck(ev.Container.ID, wksName, wksOutput)
			}
			strays = append(strays, ev.Container.ID)
			if id, ok := projectWindow(tree, strays, project); ok {
				return i3.MoveContainerToWorkspace(m.sw, id, wksName)
			}
		case <-timeout:
			tree, err := i3.Tree(m.sw)
			if err != nil {
				return err
			}
			if id, ok := projectWindow(tree, strays, project); ok {
				return i3.MoveContainerToWorkspace(m.sw, id, wksName)
			}
			return fmt.Errorf("command %q did not open a window on workspace %s after %s, it may have opened elsewhere", cfg.Command, wksName, winSpawnTimeout)
		}
	}
}

func projectWindow(tree *i3.Node, ids []int64, project string) (int64, bool) {
	base := path.Base(project)
	for _, id := range ids {
		if n := i3.FindContainer(tree, id); n != nil && strings.Contains(n.Name, base) {
			return id, true
		}
	}
	return 0, false
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
