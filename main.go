package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "select" {
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
	if len(os.Args) > 1 && os.Args[1] == "last" {
		last, err := getLast()
		if err != nil {
			log.Fatal(err)
		}
		project = last
	} else {
		selected, err := selectProject()
		if err != nil {
			log.Fatal(err)
		}

		project = projectName(selected)
	}

	leftWks, rightWks, ok, err := workspaceForProject(project)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		err := switchToWorkSpace(project, leftWks, rightWks)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err = openNew(project)
	if err != nil {
		log.Fatal(err)
	}
}

func switchToWorkSpace(name string, left, right workspace) error {
	setLast(name)
	return runCmd(
		fmt.Sprintf("workspace %s", left.Name),
		fmt.Sprintf("workspace %s", right.Name),
	)
}

func openNew(path string) error {
	leftWks, rightWks, err := nextWorkspaces()
	if err != nil {
		return err
	}
	p := projectName(path)
	setLast(p)

	err = runCmd(
		fmt.Sprintf("workspace %d", leftWks),
		fmt.Sprintf("rename workspace %d to \"%d: %s\"", leftWks, leftWks, p),
		fmt.Sprintf("exec terminator --working-directory=%s", path),
	)
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	err = runCmd(
		fmt.Sprintf("workspace %d", rightWks),
		fmt.Sprintf("rename workspace %d to \"%d: %s\"", rightWks, rightWks, p),
		fmt.Sprintf("exec zsh -c 'code %s'", path),
	)
	if err != nil {
		return err
	}

	return nil
}

func runCmd(s ...string) error {
	cmd := exec.Command("i3-msg", strings.Join(s, "; "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func projectName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		return path
	}

	resParts := []string{}
	for _, p := range parts[:len(parts)-1] {
		resParts = append(resParts, string(p[0]))
	}
	resParts = append(resParts, parts[len(parts)-1])

	return strings.Join(resParts, "/")

}
