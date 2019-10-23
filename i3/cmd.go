package i3

import (
	"os"
	"os/exec"
	"strings"
)

func cmd(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func Exec(s ...string) error {
	return cmd("i3-msg", strings.Join(s, "; "))
}
