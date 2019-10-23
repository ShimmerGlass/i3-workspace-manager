package sel

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

func Do(in io.Reader) (string, error) {
	out := &bytes.Buffer{}
	cmd := exec.Command("rofi", "-dmenu", "-sort", "-matching", "fuzzy")
	cmd.Stdin = in
	cmd.Stderr = os.Stderr
	cmd.Stdout = out

	cmd.Run()
	return strings.TrimSpace(out.String()), nil
}

func DoChoices(choices []string) (string, error) {
	out := &bytes.Buffer{}
	out.WriteString(strings.Join(choices, "\n"))

	return Do(out)
}
