package supervisor

import (
	"context"
	"sync/atomic"
)

/*
Chain supervises Components in order. All supervised components are open
in FIFO order and closed in LIFO order.

Open, Close and Wait methods may be called many times and will return equal
results. Chain guarantees that Open, Close and Wait methods for all Components
will be called once.

Chain collects and returns error from corresponding Component methods. If more
than one Components returns errors they will be wrapped in MultiError.
*/
type Chain struct {
	*compositeBase
	components []Component
}

// NewChain creates new Chain. Provided context manages whole Chain. Close
// Context is equivalent to call Chain.Close().
func NewChain(ctx context.Context, components ...Component) (c *Chain) {
	c = &Chain{
		compositeBase: newCompositeBase(ctx),
		components:    components,
	}
	return
}

// Open blocks until all components are opened in FIFO order. Chain calls
// Open() method for each descendant only after ascendant Open() returns no
// error. If Component Open() returns error Chain will close all ascendants
// in LIFO order.
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

// Close initialises shutdown for all Components in LIFO order.
func (c *Chain) Close() (err error) {
	return c.compositeBase.Close()
}

// Wait blocks until all components in Chain are exited. Chain closes each
// ascendant Component only after then its descendant Wait() method is exited.
// If one of component Wait() is exited while Chain is open all Components will
// be closed one-by-one in LIFO order.
func (c *Chain) Wait() (err error) {
	return c.compositeBase.Wait()
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
