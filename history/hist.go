package history

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
)

var histFile string

func init() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	histFile = path.Join(user.HomeDir, ".i3-wks-history")
}

func Get() ([]string, error) {
	l, err := ioutil.ReadFile(histFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(l) == 0 {
		return nil, nil
	}

	return strings.Split(string(l), "\n"), nil
}

func Add(l string) error {
	hist, err := Get()
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
	hist = append([]string{l}, hist...)

	return ioutil.WriteFile(histFile, []byte(strings.Join(hist, "\n")), 0644)
}

func Filter(projects []string) error {
	hist, err := Get()
	if err != nil {
		return err
	}
	for i := 0; i < len(hist); i++ {
		found := false
		for _, p := range projects {
			if p == hist[i] {
				found = true
				break
			}
		}
		if !found {
			copy(hist[i:], hist[i+1:])
			hist = hist[:len(hist)-1]
			break
		}
	}

	err = ioutil.WriteFile(histFile, []byte(strings.Join(hist, "\n")), 0644)
	if err != nil {
		return err
	}

	idx, err := Index()
	if err != nil {
		return err
	}

	if idx >= len(hist) {
		return SetIndex(len(hist) - 1)
	}

	return nil
}
