package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
)

func selectProject() (string, error) {
	f, err := os.Open(path.Join(home(), ".tp"))
	if err != nil {
		return "", err
	}

	selected, err := rofiSelect(f)
	if err != nil {
		return "", err
	}

	return "~/dev/" + selected, nil
}

func selectWorkspace() (string, workspace, workspace, error) {
	wks, err := workspaces()
	if err != nil {
		return "", workspace{}, workspace{}, err
	}

	p := map[string][]workspace{}
	for _, w := range wks {
		wpParts := strings.Split(w.Name, ":")
		if len(wpParts) != 2 {
			continue
		}

		wp := strings.TrimSpace(wpParts[1])
		p[wp] = append(p[wp], w)
	}

	names := []string{}
	for pn := range p {
		names = append(names, pn)
	}
	sort.Strings(names)

	out := &bytes.Buffer{}
	out.WriteString(strings.Join(names, "\n"))

	selected, err := rofiSelect(out)
	if err != nil {
		return "", workspace{}, workspace{}, err
	}

	return selected, p[selected][0], p[selected][1], nil
}

func rofiSelect(in io.Reader) (string, error) {
	out := &bytes.Buffer{}
	cmd := exec.Command("rofi", "-dmenu")
	cmd.Stdin = in
	cmd.Stderr = os.Stderr
	cmd.Stdout = out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}
