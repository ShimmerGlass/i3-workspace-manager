package main

import (
	"io/ioutil"
	"path"
)

func getLast() (string, error) {
	l, err := ioutil.ReadFile(path.Join(home(), ".i3-wks-last"))
	if err != nil {
		return "", err
	}

	return string(l), nil
}

func setLast(l string) error {
	return ioutil.WriteFile(path.Join(home(), ".i3-wks-last"), []byte(l), 0644)
}
