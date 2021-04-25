package sel

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Do(in io.Reader, prompt string, pos int) (string, error) {
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

func DoChoices(choices []string, prompt string, pos int) (string, error) {
	out := &bytes.Buffer{}
	out.WriteString(strings.Join(choices, "\n"))

	return Do(out, prompt, pos)
}
