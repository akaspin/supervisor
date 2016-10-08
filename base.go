package supervisor

import (
	"context"
	"sync"
)

type BaseComponent struct {
	context.Context
	context.CancelFunc
	wg         *sync.WaitGroup
}

func NewBaseComponent(ctx context.Context) (c *BaseComponent) {
	c = &BaseComponent{
		wg: &sync.WaitGroup{},
	}
	c.Context, c.CancelFunc = context.WithCancel(ctx)
	return
}

func (c *BaseComponent) Open() (err error) {
	c.wg.Add(1)
	go func() {
		<-c.Context.Done()
		c.wg.Done()
	}()
	return
}

func (c *BaseComponent) Close() (err error) {
	c.CancelFunc()
	return
}

func (c *BaseComponent) Wait() (err error) {
	c.wg.Wait()
	return
}
