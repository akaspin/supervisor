package supervisor

import (
	"context"
)

// Control provides ability to turn any type to supervisor component.
//
//	type MyComponent struct {
//		*Control
//	}
//
//	myComponent := &MyComponent{
//		Control: NewControl(context.Background())
//	}
//
type Control struct {
	ctx     context.Context
	cancel  context.CancelFunc
	block   *CompositeBlock
}

// NewControl returns new Control.
func NewControl(ctx context.Context, blocks ...Blocker) (c *Control) {
	c = &Control{
		block:   NewCompositeBlock(blocks...),
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return
}

func (c *Control) Open() (err error) {
	return
}

// Close closes Control context and begins evaluation of all blockers
func (c *Control) Close() (err error) {
	c.cancel()
	err = c.block.Close()
	return
}

// Wait waits for attached blockers
func (c *Control) Wait() (err error) {
	<-c.ctx.Done()
	err = c.block.Wait()
	return
}

// Ctx returns Control context. Control context is always closed before
// evaluate blockers.
func (c *Control) Ctx() context.Context {
	return c.ctx
}
