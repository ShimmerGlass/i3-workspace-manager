package history

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"
)

var histIndexFile string

func init() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	histIndexFile = path.Join(user.HomeDir, ".i3-wks-history-index")
}

func Index() (int, error) {
	l, err := ioutil.ReadFile(histIndexFile)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	if len(l) == 0 {
		return 0, nil
	}

	return strconv.Atoi(string(l))
}

func SetIndex(i int) error {
	return ioutil.WriteFile(histIndexFile, []byte(strconv.Itoa(i)), 0644)
}
