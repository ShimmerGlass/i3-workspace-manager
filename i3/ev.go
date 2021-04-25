package i3

import (
	"go.i3wm.org/i3/v4"
)

func WinEvents() (chan *i3.WindowEvent, func()) {
	winEvts := i3.Subscribe(i3.WindowEventType)
	res := make(chan *i3.WindowEvent, 100)
	go func() {
		for winEvts.Next() {
			res <- winEvts.Event().(*i3.WindowEvent)
		}
		close(res)
	}()

	return res, func() {
		winEvts.Close()
	}
}
