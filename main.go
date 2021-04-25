package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"

	"github.com/gen2brain/beeep"
)

type Manager struct {
	Workspaces     []Workspace
	SetupCommand   string
	ListCommand    string
	WksNameCommand string
}

type Workspace struct {
	Command string
	Display string
}

func parseWorkspaces(flags arrayFlags) []Workspace {
	res := []Workspace{}
	for _, f := range flags {
		i := strings.IndexByte(f, ':')
		if i < 0 {
			log.Fatalf("bad workspace flag %s", f)
		}

		res = append(res, Workspace{
			Display: f[:i],
			Command: f[i+1:],
		})
	}

	return res
}

func main() {
	logwriter, e := syslog.New(syslog.LOG_NOTICE, os.Args[0])
	if e == nil {
		log.SetOutput(logwriter)
	}

	setupCommand := flag.String("setup-cmd", "", "Setup command for project")
	listCommand := flag.String("list-cmd", "", "Setup command for listing project")
	wkNameCommand := flag.String("wksname-cmd", "", "Command to set the names of each workspace")
	var workspaces arrayFlags
	flag.Var(&workspaces, "wk", "Workspace: `display_name:command`")

	modeOpen := flag.Bool("open", false, "")
	modeSelect := flag.Bool("select", false, "")
	modePrev := flag.Bool("prev", false, "")
	modeNext := flag.Bool("next", false, "")
	modeClose := flag.Bool("close", false, "")
	flag.Parse()

	m := &Manager{
		SetupCommand:   *setupCommand,
		ListCommand:    *listCommand,
		WksNameCommand: *wkNameCommand,
		Workspaces:     parseWorkspaces(workspaces),
	}

	var err error
	switch {
	case *modeOpen:
		err = m.ActionOpen()
	case *modeSelect:
		err = m.ActionSelect()
	case *modePrev:
		err = m.ActionHistoryGo(1)
	case *modeNext:
		err = m.ActionHistoryGo(-1)
	case *modeClose:
		err = m.ActionClose()
	default:
		err = fmt.Errorf("specify one of -open, -select, -prev, -next, -close")
	}

	if err != nil {
		beeep.Notify("i3wks error", err.Error(), "")
	}
}
