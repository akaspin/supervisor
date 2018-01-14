package supervisor

import (
	"context"
)

// Control provides ability to turn any type to supervisor component
type Control struct {
	baseCtx context.Context // base context
	ctx     context.Context
	cancel  context.CancelFunc
	block   *CompositeBlock
}

// NewControl returns new Control. Control assumes provided as master. If
// provided context is closed Control.Wait returns immediately without errors.
// Control evaluates all provided blockers in CompositeBlock.
func NewControl(ctx context.Context, blocks ...Blocker) (c *Control) {
	c = &Control{
		baseCtx: ctx,
		block:   NewCompositeBlock(blocks...),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return
}

func (c *Control) Open() (err error) {
	return
}

// Close closes control and attached Blockers
func (c *Control) Close() (err error) {
	c.cancel()
	err = c.block.Close()
	return
}

// Wait waits for attached blockers or master Context
func (c *Control) Wait() (err error) {
	select {
	case <-c.baseCtx.Done():
		return
	case <-c.ctx.Done():
		err = c.block.Wait()
	}
	return
}

// Ctx returns Control context
func (c *Control) Ctx() context.Context {
	return c.ctx
}
