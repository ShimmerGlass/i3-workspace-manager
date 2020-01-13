package main

import (
	"flag"
	"fmt"
	"log"
	"path"
	"strings"
	"time"
)

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

	leftWks, rightWks, ok, err := workspaceForProject(i3Name, *leftDisplay, *rightDisplay)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		err := switchToWorkSpace(i3Name, leftWks, rightWks)
		if err != nil {
			log.Fatal(err)
		}
		return
	} else if project != "" {
		path := path.Join(*projectDirectory, project)
		err = ensureProject(path, project)
		if err != nil {
			log.Fatal(err)
		}
		err = openNew(path, project, i3Name, *leftDisplay, *rightDisplay)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func switchToWorkSpace(name string, left, right workspace) error {
	setLast(name)
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
	setLast(project)

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

func projectName(path string) string {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	return name
}
