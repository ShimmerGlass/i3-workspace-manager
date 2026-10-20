package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshuarubin/go-sway"
)

type Manager struct {
	sw         sway.Client
	nextWksNum int64

	Workspaces      []WorkspaceConfig
	SetupCommand    string
	ListCommand     string
	WksNameCommand  string
	NewCommand      string
	NewListCommand  string
	DoneCommand     string
	DoneListCommand string
}

type WorkspaceConfig struct {
	Display string
	Command string
}

func parseWorkspaces(flags arrayFlags) ([]WorkspaceConfig, error) {
	var res []WorkspaceConfig
	for _, f := range flags {
		display, command, ok := strings.Cut(f, ":")
		if !ok || display == "" || command == "" {
			return nil, fmt.Errorf("bad workspace flag %q, expected display:command", f)
		}
		res = append(res, WorkspaceConfig{Display: display, Command: command})
	}
	return res, nil
}

func main() {
	if w, err := syslog.New(syslog.LOG_NOTICE, filepath.Base(os.Args[0])); err == nil {
		log.SetOutput(w)
	}

	if err := run(); err != nil {
		log.Print(err)
		fmt.Fprintln(os.Stderr, err)
		notify("i3wks error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	setupCommand := flag.String("setup-cmd", "", "command run before opening a project")
	listCommand := flag.String("list-cmd", "", "command listing projects, one per line")
	wksNameCommand := flag.String("wksname-cmd", "", "command printing the workspace label for a project")
	newCommand := flag.String("new-cmd", "", "command creating a task from PROJECT_NAME and TASK_NAME, printing the new project")
	newListCommand := flag.String("new-list-cmd", "", "command listing projects tasks can be created from, one per line")
	doneCommand := flag.String("done-cmd", "", "command checking (I3WKS_DONE_PHASE=check) then removing (I3WKS_DONE_PHASE=remove) a task")
	doneListCommand := flag.String("done-list-cmd", "", "command listing task projects, one per line")
	var workspaces arrayFlags
	flag.Var(&workspaces, "wk", "workspace `display:command`, display is left, right or an output name (repeatable)")

	modeProject := flag.String("project", "", "open the named project, without the selector")
	modeOpen := flag.Bool("open", false, "open a project from the project list")
	modeSelect := flag.Bool("select", false, "switch to an already opened project")
	modePrev := flag.Bool("prev", false, "switch to the previous project")
	modeNext := flag.Bool("next", false, "switch to the next project")
	modeClose := flag.Bool("close", false, "close projects")
	modeNew := flag.Bool("new", false, "create a task project and open it")
	modeDone := flag.Bool("done", false, "close a task project and remove it")
	modeDoneMerged := flag.Bool("done-merged", false, "close and remove every task whose PR is merged, without asking")
	modeDaemon := flag.Bool("daemon", false, "keep project workspaces pinned to their display")
	flag.Parse()

	wks, err := parseWorkspaces(workspaces)
	if err != nil {
		return err
	}

	sw, err := sway.New(context.Background())
	if err != nil {
		return err
	}

	m := &Manager{
		sw:              sw,
		Workspaces:      wks,
		SetupCommand:    *setupCommand,
		ListCommand:     *listCommand,
		WksNameCommand:  *wksNameCommand,
		NewCommand:      *newCommand,
		NewListCommand:  *newListCommand,
		DoneCommand:     *doneCommand,
		DoneListCommand: *doneListCommand,
	}

	switch {
	case *modeProject != "":
		return m.ActionOpenProject(*modeProject)
	case *modeOpen:
		return m.ActionOpen()
	case *modeSelect:
		return m.ActionSelect()
	case *modePrev:
		return m.ActionHistoryGo(1)
	case *modeNext:
		return m.ActionHistoryGo(-1)
	case *modeClose:
		return m.ActionClose()
	case *modeNew:
		return m.ActionNew()
	case *modeDone:
		return m.ActionDone()
	case *modeDoneMerged:
		return m.ActionDoneMerged()
	case *modeDaemon:
		return m.ActionDaemon()
	default:
		return errors.New("specify one of -project, -open, -select, -prev, -next, -close, -new, -done, -done-merged, -daemon")
	}
}
