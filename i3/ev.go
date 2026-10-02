package i3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joshuarubin/go-sway"
)

const subscribeTimeout = 2 * time.Second

type WindowEvent = sway.WindowEvent

type winHandler struct {
	sway.EventHandler
	ctx    context.Context
	events chan WindowEvent
	ready  chan struct{}
}

func (h *winHandler) Tick(_ context.Context, ev sway.TickEvent) {
	if ev.First {
		close(h.ready)
	}
}

func (h *winHandler) Window(_ context.Context, ev WindowEvent) {
	select {
	case h.events <- ev:
	case <-h.ctx.Done():
	}
}

func WinEvents() (<-chan WindowEvent, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	h := &winHandler{
		EventHandler: sway.NoOpEventHandler(),
		ctx:          ctx,
		events:       make(chan WindowEvent, 100),
		ready:        make(chan struct{}),
	}

	errc := make(chan error, 1)
	go func() {
		errc <- sway.Subscribe(ctx, h, sway.EventTypeWindow, sway.EventTypeTick)
	}()

	select {
	case <-h.ready:
		return h.events, cancel, nil
	case err := <-errc:
		cancel()
		return nil, nil, fmt.Errorf("subscribing to window events: %w", err)
	case <-time.After(subscribeTimeout):
		cancel()
		return nil, nil, errors.New("timed out subscribing to window events")
	}
}

func IsNewTiledWindow(ev WindowEvent) bool {
	return ev.Change == sway.WindowNew && ev.Container.Type != sway.NodeFloatingCon
}
