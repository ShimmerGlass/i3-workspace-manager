package main

import (
	"flag"
	"fmt"
	"log"
	"path"
	"time"
)

type Manager struct {
	leftDisplay  string
	rightDisplay string
	projectDir   string
}

func main() {
	projectDirectory := flag.String("project-dir", "", "")
	leftDisplay := flag.String("left-display", "", "")
	rightDisplay := flag.String("right-display", "", "")
	modeSelect := flag.Bool("select", false, "")
	modeLast := flag.Bool("last", false, "")
	flag.Parse()

	if *modeSelect {
		selected, leftWks, rightWks, err := selectWorkspace()
		if err != nil {
			log.Fatal(err)
		}

		err = switchToWorkSpace(selected, leftWks, rightWks)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	var project string
	var i3Name string
	if *modeLast {
		last, err := getLast()
		if err != nil {
			log.Fatal(err)
		}
		i3Name = last
	} else {
		selected, err := selectProject()
		if err != nil {
			log.Fatal(err)
		}

		project = selected
		i3Name = projectName(selected)
	}

}

func goBack() error {
	hist, err := getHistory()
	if err != nil {
		return err
	}

	current, err := currentProject()
	if err != nil {

	}
	for i := len(hist) - 1; i >= 0; i-- {
		if hist[i] == current {
			continue
		}

	}
}

func (m *Manager) openProject(project string) error {
	leftWks, rightWks, ok, err := workspaceForProject(project, leftDisplay, rightDisplay)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		err := switchToWorkSpace(project, leftWks, rightWks)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	} else if project != "" {
		path := path.Join(*projectDirectory, project)
		err = ensureProject(path, project)
		if err != nil {
			log.Fatal(err)
		}
		err = openNew(path, project, i3Name, leftDisplay, rightDisplay)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func switchToWorkSpace(name string, left, right workspace) error {
	addToHistory(name)
	return i3Cmd(
		fmt.Sprintf("workspace %s", left.Name),
		fmt.Sprintf("workspace %s", right.Name),
	)
}

func openNew(path, project, i3Name, leftDisplay, rightDisplay string) error {

	leftWks, rightWks, err := nextWorkspaces(leftDisplay, rightDisplay)
	if err != nil {
		return err
	}
	addToHistory(project)

	err = i3Cmd(
		fmt.Sprintf("workspace %d", leftWks),
		fmt.Sprintf("rename workspace %d to \"%d: %s\"", leftWks, leftWks, project),
		fmt.Sprintf("exec terminator --working-directory=%s", path),
	)
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	err = i3Cmd(
		fmt.Sprintf("workspace %d", rightWks),
		fmt.Sprintf("rename workspace %d to \"%d: %s\"", rightWks, rightWks, project),
	)
	if err != nil {
		return err
	}

	cmd("zsh", "-c", fmt.Sprintf("source ~/.zshrc; code %s", path))

	return nil
}
