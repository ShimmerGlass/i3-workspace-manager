package i3

import (
	"context"

	"github.com/joshuarubin/go-sway"
)

var _ sway.EventHandler = &swayEventHandler{}

type swayEventHandler struct {
	winCb func(sway.WindowEvent)
}

// BarConfigUpdate implements [sway.EventHandler].
func (e *swayEventHandler) BarConfigUpdate(context.Context, sway.BarConfigUpdateEvent) {

}

// BarStateUpdate implements [sway.EventHandler].
func (e *swayEventHandler) BarStateUpdate(context.Context, sway.BarStateUpdateEvent) {

}

// BarStatusUpdate implements [sway.EventHandler].
func (e *swayEventHandler) BarStatusUpdate(context.Context, sway.BarStatusUpdateEvent) {

}

// Binding implements [sway.EventHandler].
func (e *swayEventHandler) Binding(context.Context, sway.BindingEvent) {

}

// Input implements [sway.EventHandler].
func (e *swayEventHandler) Input(context.Context, sway.InputEvent) {

}

// Mode implements [sway.EventHandler].
func (e *swayEventHandler) Mode(context.Context, sway.ModeEvent) {

}

// Shutdown implements [sway.EventHandler].
func (e *swayEventHandler) Shutdown(context.Context, sway.ShutdownEvent) {

}

// Tick implements [sway.EventHandler].
func (e *swayEventHandler) Tick(context.Context, sway.TickEvent) {

}

// Window implements [sway.EventHandler].
func (e *swayEventHandler) Window(_ context.Context, ev sway.WindowEvent) {
	e.winCb(ev)
}

// Workspace implements [sway.EventHandler].
func (e *swayEventHandler) Workspace(context.Context, sway.WorkspaceEvent) {

}

func WinEvents() (chan sway.WindowEvent, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())

	res := make(chan sway.WindowEvent, 100)

	eh := &swayEventHandler{
		winCb: func(we sway.WindowEvent) {
			res <- we
		},
	}

	go func() {
		sway.Subscribe(ctx, eh, sway.EventTypeWindow)
	}()

	return res, cancel, nil
}
