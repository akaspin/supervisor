package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Blocker
type Blocker interface {
	// Close begins blocker evaluation
	Close() (err error)

	// Wait waits for blocker
	Wait() (err error)
}

// CompositeBlock evaluates multiple blocks
type CompositeBlock struct {
	blocks []Blocker
}

// NewCompositeBlock returns new composite block
func NewCompositeBlock(blocks ...Blocker) (c *CompositeBlock) {
	c = &CompositeBlock{
		blocks: blocks,
	}
	return
}

// Close calls Close method of all blocks in Composite blocker
func (c *CompositeBlock) Close() (err error) {
	for _, block := range c.blocks {
		if err1 := block.Close(); err1 != nil {
			err = AppendError(err, err1)
		}
	}
	return
}

// Wait wait for all blocks or first error from block
func (c *CompositeBlock) Wait() (err error) {
	if len(c.blocks) == 0 {
		return
	}
	cond := &sync.Cond{
		L: &sync.Mutex{},
	}
	var done int

	for _, block := range c.blocks {
		go func(block Blocker) {
			blockErr := block.Wait()
			cond.L.Lock()
			done++
			if blockErr != nil {
				err = AppendError(err, blockErr)
			}
			cond.L.Unlock()
			cond.Broadcast()
		}(block)
	}
	cond.L.Lock()
	for done != len(c.blocks) && err == nil {
		cond.Wait()
	}
	cond.L.Unlock()
	return
}

// TimeoutBlock blocks until provided Context is closed or timeout reached
type TimeoutBlock struct {
	baseCtx context.Context
	timeout time.Duration

	closed   int32
	failChan chan struct{}
}

func NewTimeoutBlock(ctx context.Context, timeout time.Duration) (b *TimeoutBlock) {
	b = &TimeoutBlock{
		baseCtx:  ctx,
		timeout:  timeout,
		failChan: make(chan struct{}),
	}
	return
}

func (b *TimeoutBlock) Close() (err error) {
	if !atomic.CompareAndSwapInt32(&b.closed, 0, 1) {
		return
	}
	go func() {
		time.Sleep(b.timeout)
		close(b.failChan)
	}()
	return
}

func (b *TimeoutBlock) Wait() (err error) {
	select {
	case <-b.baseCtx.Done():
		return nil
	case <-b.failChan:
		err = context.DeadlineExceeded
	}
	return err
}
