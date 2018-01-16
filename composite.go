package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
)

type compositeBase struct {
	ctx    context.Context
	cancel context.CancelFunc

	opened uint32
	openE  componentErr

	closeWg sync.WaitGroup
	closeE  componentErr

	waitWg sync.WaitGroup
	waitE  componentErr
}

func newCompositeBase(ctx context.Context) (b *compositeBase) {
	b = &compositeBase{}
	b.ctx, b.cancel = context.WithCancel(ctx)
	return b
}

func (b *compositeBase) Close() (err error) {
	if atomic.LoadUint32(&b.opened) == 0 {
		return ErrNotOpened
	}
	b.cancel()
	b.closeWg.Wait()
	err = b.closeE.get()
	return err
}

func (c *compositeBase) Wait() (err error) {
	if atomic.LoadUint32(&c.opened) == 0 {
		return ErrNotOpened
	}
	c.waitWg.Wait()
	err = c.waitE.get()
	return err
}
