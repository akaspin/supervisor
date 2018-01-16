package supervisor

import (
	"context"
	"sync/atomic"
)

type Chain1 struct {
	*compositeBase
	components []Component
}

func NewChain1(ctx context.Context, components ...Component) (c *Chain1) {
	c = &Chain1{
		compositeBase: newCompositeBase(ctx),
		components:    components,
	}
	return
}

func (c *Chain1) Open() (err error) {
	if !atomic.CompareAndSwapUint32(&c.opened, 0, 1) {
		return c.openE.get()
	}

	if len(c.components) == 0 {
		c.openE.set(ErrEmptyComposite)
		err = ErrEmptyComposite
		return
	}
	_, cancel := context.WithCancel(context.Background())
	if err = c.build(cancel, c.components); err != nil {
		c.openE.set(err)
	}
	return
}

func (c *Chain1) build(ascendantCancel context.CancelFunc, tail []Component) (err error) {
	if len(tail) == 0 {
		// supervise last chunk
		go func() {
			<-c.ctx.Done()
			ascendantCancel()
		}()
		return nil
	}
	component := tail[0]
	ctx, cancel := context.WithCancel(context.Background())

	if openErr := component.Open(); openErr != nil {
		err = AppendError(err, openErr)
		ascendantCancel()
		//c.cancel()
		return err
	}
	// supervise close
	c.closeWg.Add(1)
	go func() {
		defer c.closeWg.Done()
		<-ctx.Done()
		if closeErr := component.Close(); closeErr != nil {
			c.closeE.set(closeErr)
		}
	}()
	// supervise wait
	c.waitWg.Add(1)
	go func() {
		defer c.waitWg.Done()
		if waitErr := component.Wait(); waitErr != nil {
			c.waitE.set(waitErr)
		}
		select {
		case <-c.ctx.Done():
			// normal shutdown
			ascendantCancel()
		default:
			// abnormal shutdown we need close context and wait for descendants
			c.cancel()
			<-ctx.Done()
			ascendantCancel()
		}
	}()

	err = c.build(cancel, tail[1:])
	return nil
}

type chainComponent struct {
	ascendant    *chainComponent
	component    Component
	shutdownChan chan struct{} // upstream close
}
