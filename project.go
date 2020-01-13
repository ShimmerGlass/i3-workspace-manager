package main

import (
	"os"
	"path/filepath"
)

func ensureProject(path, name string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	err = cmd("git", "clone", "ssh://review.criteois.lan:29418/"+name, path)
	if err != nil {
		return err
	}
	err = cmd("scp", "-p", "-P", "29418", "review.criteois.lan:hooks/commit-msg", filepath.Join(path, ".git/hooks/"))
	if err != nil {
		return err
	}

	return nil
}
