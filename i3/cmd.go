package i3

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func cmd(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	err := cmd.Run()
	if err != nil && out.Len() > 0 {
		return fmt.Errorf("error running %s %v: %s: %s", bin, args, err, out.String())
	} else if err != nil {
		return fmt.Errorf("error running %s %v: %s", bin, args, err)
	}
	return nil
}

func Exec(s ...string) error {
	return cmd("i3-msg", strings.Join(s, "; "))
}
