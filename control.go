package supervisor

import (
	"context"
	"sync"
)

// Control provides ability to turn any type to supervisor component
type Control struct {
	ctx    context.Context

	// Cancel cancels control context
	Cancel context.CancelFunc

	// internal wait group
	wg     *sync.WaitGroup

	// bounded wait group
	boundedWg *sync.WaitGroup
}

func NewControl(ctx context.Context) (c *Control) {
	c = &Control{
		wg: &sync.WaitGroup{},
		boundedWg: &sync.WaitGroup{},
	}
	c.ctx, c.Cancel = context.WithCancel(ctx)
	return
}

func (c *Control) Open() (err error) {
	c.wg.Add(1)
	go func() {
		<-c.ctx.Done()
		c.boundedWg.Wait()
		c.wg.Done()
	}()
	return
}

func (c *Control) Close() (err error) {
	c.Cancel()
	return
}

func (c *Control) Wait() (err error) {
	c.wg.Wait()
	return
}

// Ctx returns Control context
func (c *Control) Ctx() context.Context {
	return c.ctx
}

// Acquire increases internal lock counter
func (c *Control) Acquire() {
	c.boundedWg.Add(1)
}

func (c *Control) Release() {
	c.boundedWg.Done()
}

