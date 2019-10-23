# i3-workspace-manager

When working on a lot of projects at one managing workspaces and windows can become cumbersome.
*i3-workspace-manager* is a way to automatically open a text editor and a terminal from a project list and switch between them.
It takes a command to list your projects, and a list of command to run on your displays to open it.

The following commands are available :

| Command         | Description                                          |
|-----------------|------------------------------------------------------|
| `-open`         | open a new project from project list                 |
| `-select`       | switch to an already opened project                  |
| `-prev / -next` | quick switch to previous / next project              |
| `-close`        | close project windows and switch to previous project |

## Project list

Listing all git repositories in ~/dev:
```
-list-cmd 'cd ~/dev && find . -type d -name .git | sed "s/.git$//" | sed "s/^\.\///"'
```

## Workspaces

Workspace command are given $PROJECT_NAME env var as provided by the `-list-cmd`

For example open a terminal on display 1:
```
-wk 'DVI-D-0:terminator --working-directory ~/dev/$PROJECT_NAME'
```
And an editor on display 2 :
```
-wk 'DVI-I-1:code ~/dev/$PROJECT_NAME'
```

## Setup command

`-setup-cmd` will be called before opening a project, as with `-wk`, $PROJECT_NAME is given as env var. This allows to perform some action such as cloning the repository if it isn't already.

## i3 config example

In this example the commands are moved to script file for easy change without reloading i3

```
# Common command for all actions
set $i3wks ~/dev/go/bin/i3-workspace-manager -wk 'DVI-D-0:~/bin/i3wks-cmd' -wk 'DVI-I-1:~/bin/i3wks-code' -setup-cmd '~/bin/i3wks-setup' -list-cmd '~/bin/i3wks-list'

# Select action
bindsym $mod+n exec --no-startup-id $i3wks -select

# Open action
bindsym $mod+Shift+n exec --no-startup-id $i3wks -open

# Prev action
bindsym $mod+b exec --no-startup-id $i3wks -prev

# Next action
bindsym $mod+Shift+b exec --no-startup-id $i3wks -next

# Close action
bindsym $mod+Shift+v exec --no-startup-id $i3wks -close
```

## Dependencies

* [Rofi](https://github.com/davatorium/rofi) for displaying the project switcher.