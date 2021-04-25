package sel

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func DoChoices(choices []string, prompt string, pos int) (string, error) {
	in := &bytes.Buffer{}
	in.WriteString(strings.Join(choices, "\n"))

	out := &bytes.Buffer{}
	cmd := exec.Command(
		"rofi",
		"-dmenu",
		"-matching", "fuzzy",
		"-mesg", prompt,
		"-selected-row", strconv.Itoa(pos),
	)
	cmd.Stdin = in
	cmd.Stderr = os.Stderr
	cmd.Stdout = out

	cmd.Run()
	return strings.TrimSpace(out.String()), nil
}

func DoChoicesMulti(choices []string, prompt string, pos int) ([]string, error) {
	in := &bytes.Buffer{}
	in.WriteString(strings.Join(choices, "\n"))

	out := &bytes.Buffer{}
	cmd := exec.Command(
		"rofi",
		"-dmenu",
		"-matching", "fuzzy",
		"-mesg", prompt,
		"-selected-row", strconv.Itoa(pos),
		"-multi-select",
	)
	cmd.Stdin = in
	cmd.Stderr = os.Stderr
	cmd.Stdout = out

	cmd.Run()

	res := []string{}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		res = append(res, strings.TrimSpace(scanner.Text()))
	}
	return res, nil
}
