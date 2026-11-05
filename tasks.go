package main

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/shimmerglass/i3-workspace-manager/history"
	"github.com/shimmerglass/i3-workspace-manager/sel"
)

func (m *Manager) ActionNew() error {
	if m.NewCommand == "" || m.NewListCommand == "" {
		return errors.New("-new-cmd and -new-list-cmd are required")
	}

	base, err := m.pickTaskBase()
	if err != nil || base == "" {
		return err
	}

	name, err := sel.Prompt(fmt.Sprintf("New task from %s", base))
	if err != nil || name == "" {
		return err
	}

	out, err := runShell(m.NewCommand, append(projectEnv(base), "TASK_NAME="+name))
	if err != nil {
		return fmt.Errorf("creating task %s: %w", name, err)
	}
	project := strings.TrimSpace(out)
	if project == "" {
		return fmt.Errorf("creating task %s: -new-cmd printed no project", name)
	}

	hist, err := history.Get()
	if err != nil {
		return err
	}
	return m.openAndRecord(hist, project)
}

func (m *Manager) pickTaskBase() (string, error) {
	bases, err := listProjects(m.NewListCommand)
	if err != nil {
		return "", err
	}
	switch len(bases) {
	case 0:
		return "", errors.New("-new-list-cmd listed no project")
	case 1:
		return bases[0], nil
	}

	hist, err := history.Get()
	if err != nil {
		return "", err
	}
	hist.Sort(bases)
	return sel.DoChoices(bases, "New task from", 0)
}

func (m *Manager) ActionDone() error {
	if m.DoneCommand == "" || m.DoneListCommand == "" {
		return errors.New("-done-cmd and -done-list-cmd are required")
	}

	tasks, err := listProjects(m.DoneListCommand)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		notify("i3wks", "No task to mark done")
		return nil
	}

	hist, err := history.Get()
	if err != nil {
		return err
	}
	hist.Sort(tasks)

	current, _, err := m.CurrentProject()
	if err != nil {
		return err
	}
	pos := max(slices.Index(tasks, current), 0)

	choices := append([]string{allMergedChoice}, tasks...)
	project, err := sel.DoChoices(choices, "Task done", pos+1)
	if err != nil || project == "" {
		return err
	}
	if project == allMergedChoice {
		return m.doneAllMerged(hist, tasks, true)
	}

	out, err := runShell(m.DoneCommand, doneEnv(project, "check"))
	if err != nil {
		return fmt.Errorf("checking task %s: %w", project, err)
	}
	if warn := strings.TrimSpace(out); warn != "" {
		ok, err := sel.Confirm(fmt.Sprintf("%s\n\nDelete %s anyway?", warn, project))
		if err != nil || !ok {
			return err
		}
	}

	if err := m.removeTasks(hist, []string{project}); err != nil {
		return err
	}
	notify("i3wks", fmt.Sprintf("Task %s done", project))
	return nil
}

func (m *Manager) ActionDoneTask(project string, force bool) error {
	if m.DoneCommand == "" || m.DoneListCommand == "" {
		return errors.New("-done-cmd and -done-list-cmd are required")
	}

	tasks, err := listProjects(m.DoneListCommand)
	if err != nil {
		return err
	}
	if !slices.Contains(tasks, project) {
		return fmt.Errorf("%s is not a task listed by -done-list-cmd", project)
	}

	out, err := runShell(m.DoneCommand, doneEnv(project, "check"))
	if err != nil {
		return fmt.Errorf("checking task %s: %w", project, err)
	}
	if warn := strings.TrimSpace(out); warn != "" && !force {
		return fmt.Errorf("%s\nnot deleting %s, pass -force to delete it anyway", warn, project)
	}

	hist, err := history.Get()
	if err != nil {
		return err
	}
	if err := m.removeTasks(hist, []string{project}); err != nil {
		return err
	}
	notify("i3wks", fmt.Sprintf("Task %s done", project))
	return nil
}

func (m *Manager) ActionDoneMerged() error {
	if m.DoneCommand == "" || m.DoneListCommand == "" {
		return errors.New("-done-cmd and -done-list-cmd are required")
	}

	tasks, err := listProjects(m.DoneListCommand)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		notify("i3wks", "No task to mark done")
		return nil
	}

	hist, err := history.Get()
	if err != nil {
		return err
	}
	return m.doneAllMerged(hist, tasks, false)
}

const allMergedChoice = "✗ all merged tasks"

func (m *Manager) doneAllMerged(hist *history.History, tasks []string, confirm bool) error {
	safe := make([]bool, len(tasks))
	var wg sync.WaitGroup
	for i, task := range tasks {
		wg.Go(func() {
			out, err := runShell(m.DoneCommand, doneEnv(task, "check-merged"))
			if err != nil {
				log.Printf("checking task %s: %s", task, err)
				return
			}
			safe[i] = strings.TrimSpace(out) == ""
		})
	}
	wg.Wait()

	var merged []string
	for i, task := range tasks {
		if safe[i] {
			merged = append(merged, task)
		}
	}
	if len(merged) == 0 {
		notify("i3wks", "No merged task to delete")
		return nil
	}

	if confirm {
		ok, err := sel.Confirm(fmt.Sprintf("Delete %d merged task(s)?\n\n%s", len(merged), strings.Join(merged, "\n")))
		if err != nil || !ok {
			return err
		}
	}

	if err := m.removeTasks(hist, merged); err != nil {
		return err
	}
	notify("i3wks", fmt.Sprintf("%d merged task(s) deleted\n\n%s", len(merged), strings.Join(merged, "\n")))
	return nil
}

func (m *Manager) removeTasks(hist *history.History, projects []string) error {
	current, _, err := m.CurrentProject()
	if err != nil {
		return err
	}
	wks, err := m.managedWorkspaces()
	if err != nil {
		return err
	}

	var errs []error
	for _, project := range projects {
		if err := m.closeProject(wks, project); err != nil {
			errs = append(errs, err)
			continue
		}
		if _, err := runShell(m.DoneCommand, doneEnv(project, "remove")); err != nil {
			errs = append(errs, fmt.Errorf("removing task %s: %w", project, err))
			continue
		}
		hist.Remove(project)
	}

	if projects, err := m.ProjectList(); err != nil {
		log.Printf("pruning history: %s", err)
	} else if len(projects) > 0 {
		hist.Retain(projects)
	}
	if err := history.Write(hist); err != nil {
		log.Printf("writing history: %s", err)
	}
	if slices.Contains(projects, current) {
		errs = append(errs, m.returnToRecent(hist))
	}
	return errors.Join(errs...)
}

func doneEnv(project, phase string) []string {
	return append(projectEnv(project), "I3WKS_DONE_PHASE="+phase)
}
