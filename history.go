package main

import (
	"io/ioutil"
	"path"
	"strings"
)

func getHistory() ([]string, error) {
	l, err := ioutil.ReadFile(path.Join(home(), ".i3-wks-history"))
	if err != nil {
		return nil, err
	}

	return strings.Split(string(l), "\n"), nil
}

func addToHistory(l string) error {
	hist, err := getHistory()
	if err != nil {
		return err
	}
	for i := 0; i < len(hist); i++ {
		if hist[i] == l {
			copy(hist[i:], hist[i+1:])
			hist = hist[:len(hist)-1]
			break
		}
	}
	hist = append(hist, l)

	return ioutil.WriteFile(path.Join(home(), ".i3-wks-last"), []byte(strings.Join(hist, "\n")), 0644)
}
