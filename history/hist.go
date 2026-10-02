package history

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	maxSize  = 200
	fileName = ".i3-wks-history"
)

type History struct {
	Position int
	Projects []string
}

func (h *History) Remove(p string) {
	h.Projects = slices.DeleteFunc(h.Projects, func(x string) bool { return x == p })
}

func (h *History) Add(p string) {
	h.Remove(p)
	h.Projects = slices.Insert(h.Projects, 0, p)
	h.Position = 0
}

func (h *History) Sort(s []string) {
	idx := make(map[string]int, len(h.Projects))
	for i, p := range h.Projects {
		idx[p] = i
	}
	rank := func(p string) int {
		if i, ok := idx[p]; ok {
			return i
		}
		return len(h.Projects)
	}

	slices.SortStableFunc(s, func(a, b string) int {
		return cmp.Compare(rank(a), rank(b))
	})
}

func filePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, fileName), nil
}

func Get() (*History, error) {
	p, err := filePath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(p)
	if errors.Is(err, fs.ErrNotExist) {
		return &History{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parse(f)
}

func Write(h *History) error {
	p, err := filePath()
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(p), fileName+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.Write(h.format())
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), p)
}

func parse(r io.Reader) (*History, error) {
	res := &History{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 || (line[0] != '>' && line[0] != ' ') {
			continue
		}

		res.Projects = append(res.Projects, strings.TrimSpace(line[2:]))
		if line[0] == '>' {
			res.Position = len(res.Projects) - 1
		}
	}

	return res, scanner.Err()
}

func (h *History) format() []byte {
	projects := h.Projects[:min(len(h.Projects), maxSize)]
	pos := min(h.Position, len(projects)-1)

	buf := &bytes.Buffer{}
	for i, p := range projects {
		if i == pos {
			buf.WriteString("> ")
		} else {
			buf.WriteString("  ")
		}
		buf.WriteString(p)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}
