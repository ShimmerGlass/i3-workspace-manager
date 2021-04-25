package history

import (
	"bufio"
	"bytes"
	"math"
	"os"
	"os/user"
	"path"
	"sort"
	"strings"
)

const historyMaxSize = 200

var histFile string

func init() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	histFile = path.Join(user.HomeDir, ".i3-wks-history")
}

type History struct {
	Position int
	Projects []string
}

func (h *History) Remove(p string) {
	for i := range h.Projects {
		if h.Projects[i] != p {
			continue
		}

		copy(h.Projects[i:], h.Projects[i+1:])
		h.Projects = h.Projects[:len(h.Projects)-1]
	}
}

func (h *History) Add(p string) {
	h.Remove(p)
	h.Projects = append([]string{p}, h.Projects...)
	h.Position = 0
}

func (h *History) Sort(s []string) {
	sort.SliceStable(s, func(i, j int) bool {
		ip := h.projectIdx(s[i])
		jp := h.projectIdx(s[j])

		return ip < jp
	})
}

func (h *History) projectIdx(p string) int {
	for i, hp := range h.Projects {
		if hp == p {
			return i
		}
	}

	return math.MaxInt32
}

func Get() (*History, error) {
	res := &History{}

	f, err := os.Open(histFile)
	if os.IsNotExist(err) {
		return res, nil
	}
	if err != nil {
		return res, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}

		if line[0] != '>' && line[0] != ' ' {
			continue
		}

		res.Projects = append(res.Projects, strings.TrimSpace(line[2:]))
		if line[0] == '>' {
			res.Position = len(res.Projects) - 1
		}
	}

	if scanner.Err() != nil {
		return res, scanner.Err()
	}

	return res, nil

}

func Write(h *History) error {
	buf := &bytes.Buffer{}

	projects := h.Projects
	if len(projects) > historyMaxSize {
		projects = projects[:historyMaxSize]
	}

	for i, p := range projects {
		if h.Position == i {
			buf.WriteByte('>')
		} else {
			buf.WriteByte(' ')
		}
		buf.WriteByte(' ')
		buf.WriteString(p)
		buf.WriteByte('\n')
	}

	return os.WriteFile(histFile, buf.Bytes(), 0o644)
}
