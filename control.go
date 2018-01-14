package supervisor

import (
	"context"
	"errors"
)

var (
	ErrCloseTimeoutExceeded = errors.New("close timeout exceeded")
)

// Control provides ability to turn any type to supervisor component
type Control struct {
	ctx    context.Context
	cancel context.CancelFunc

	closeCtx    context.Context
	closeCancel context.CancelFunc
}

func NewControl(ctx context.Context) (c *Control) {
	c = &Control{}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.closeCtx, c.closeCancel = context.WithCancel(context.Background())
	return
}

func (c *Control) Open() (err error) {
	go func() {
		<-c.ctx.Done()
		c.closeCancel()
	}()

	return
}

func (c *Control) Close() (err error) {
	c.cancel()
	return
}

func (c *Control) Wait() (err error) {
	<-c.closeCtx.Done()
	return
}

// Ctx returns Control context
func (c *Control) Ctx() context.Context {
	return c.ctx
}
