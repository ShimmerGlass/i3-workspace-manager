package sel

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func DoChoices(choices []string, prompt string, pos int) (string, error) {
	out, err := rofi(choices, prompt, pos)
	return strings.TrimSpace(out), err
}

func DoChoicesMulti(choices []string, prompt string, pos int) ([]string, error) {
	out, err := rofi(choices, prompt, pos, "-multi-select")
	if err != nil {
		return nil, err
	}

	var res []string
	for line := range strings.Lines(out) {
		if line = strings.TrimSpace(line); line != "" {
			res = append(res, line)
		}
	}
	return res, nil
}

func rofi(choices []string, prompt string, pos int, extra ...string) (string, error) {
	mesg, placeholder := splitPrompt(prompt)
	args := []string{
		"-dmenu",
		"-matching", "fuzzy",
		"-theme-str", fmt.Sprintf("entry { placeholder: %s; }", rasiString(placeholder)),
		"-selected-row", strconv.Itoa(pos),
	}
	if mesg != "" {
		args = append(args, "-mesg", mesg)
	}
	args = append(args, extra...)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("rofi", args...)
	cmd.Stdin = strings.NewReader(strings.Join(choices, "\n"))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok && exitErr.ExitCode() == 1 {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("rofi: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func Prompt(prompt string) (string, error) {
	return DoChoices(nil, prompt, 0)
}

func Confirm(prompt string) (bool, error) {
	out, err := DoChoices([]string{"no", "yes"}, prompt, 0)
	return out == "yes", err
}

func splitPrompt(prompt string) (mesg, placeholder string) {
	prompt = strings.TrimSpace(prompt)
	i := strings.LastIndexByte(prompt, '\n')
	if i < 0 {
		return "", prompt
	}
	return strings.TrimSpace(prompt[:i]), strings.TrimSpace(prompt[i+1:])
}

func rasiString(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
