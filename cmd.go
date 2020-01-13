package main

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

func i3Cmd(s ...string) error {
	return cmd("i3-msg", strings.Join(s, "; "))
}
