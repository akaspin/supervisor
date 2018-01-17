package supervisor

import (
	"context"
	"sync/atomic"
)

type Chain struct {
	*compositeBase
	components []Component
}

func NewChain(ctx context.Context, components ...Component) (c *Chain) {
	c = &Chain{
		compositeBase: newCompositeBase(ctx),
		components:    components,
	}
	return
}

func (c *Chain) Open() (err error) {
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

func (c *Chain) build(ascendantCancel context.CancelFunc, tail []Component) (err error) {
	if len(tail) == 0 {
		// supervise last chunk
		go func() {
			<-c.ctx.Done()
			ascendantCancel()
		}()
		return nil
	}
	component := tail[0]

	if openErr := component.Open(); openErr != nil {
		err = AppendError(err, openErr)
		ascendantCancel()
		c.cancel()
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	var waitExited uint32

	// supervise close
	c.closeWg.Add(1)
	go func() {
		defer c.closeWg.Done()
		<-ctx.Done()
		if atomic.CompareAndSwapUint32(&waitExited, 0, 1) {
			if closeErr := component.Close(); closeErr != nil {
				c.closeE.set(closeErr)
			}
		}
	}()

	// supervise wait
	c.waitWg.Add(1)
	go func() {
		defer c.waitWg.Done()
		if waitErr := component.Wait(); waitErr != nil {
			c.waitE.set(waitErr)
		}
		atomic.CompareAndSwapUint32(&waitExited, 0, 1)
		select {
		case <-c.ctx.Done():
			// normal shutdown
		default:
			// abnormal shutdown we need close context and wait for descendants
			c.cancel()
			<-ctx.Done()
		}
		ascendantCancel()
		cancel()
	}()

	return c.build(cancel, tail[1:])
}

type chainComponent struct {
	ascendant    *chainComponent
	component    Component
	shutdownChan chan struct{} // upstream close
}
