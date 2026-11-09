package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func projectShell(project, command string) string {
	return "PROJECT_NAME=" + shellQuote(project) + " exec /bin/sh -c " + shellQuote(command)
}

func swayExec(script string) string {
	payload := base64.StdEncoding.EncodeToString([]byte(script))
	return `exec /bin/sh -c 'eval "$(printf %s ` + payload + ` | base64 -d)"'`
}

func checkSwayName(s string) error {
	if strings.ContainsAny(s, "\n\r\x00") {
		return fmt.Errorf("%q contains a character sway cannot carry in a command", s)
	}
	return nil
}

func buildSpawnCommand(wksName, assignOutput, project, command string, restore []string) (string, error) {
	for _, s := range append([]string{wksName, assignOutput}, restore...) {
		if err := checkSwayName(s); err != nil {
			return "", err
		}
	}

	var parts []string
	if assignOutput != "" {
		parts = append(parts, "workspace "+i3.Quote(wksName)+" output "+i3.Quote(assignOutput))
	}
	parts = append(parts, "workspace --no-auto-back-and-forth "+i3.Quote(wksName))
	parts = append(parts, swayExec(projectShell(project, command)))
	for _, name := range restore {
		parts = append(parts, "workspace --no-auto-back-and-forth "+i3.Quote(name))
	}
	return strings.Join(parts, "; "), nil
}

func restoreTargets(current []i3.Workspace, wksOutput string) []string {
	var focused *i3.Workspace
	for i := range current {
		if current[i].Focused {
			focused = &current[i]
			break
		}
	}
	if focused == nil {
		return nil
	}

	var res []string
	if wksOutput != "" && wksOutput != focused.Output {
		for _, w := range current {
			if w.Output == wksOutput && w.Visible {
				res = append(res, w.Name)
			}
		}
	}
	return append(res, focused.Name)
}
