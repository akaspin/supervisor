package supervisor

import (
	"context"
	"time"
)

// Blocker blocks Control Wait
type Blocker interface {

	// Block begins blocker evaluation
	Block() (err error)

	// Wait blocks until
	Wait() (err error)
}

// TimeoutBlock blocks until provided Context is closed or timeout reached
type TimeoutBlock struct {
	baseCtx context.Context
	timeout time.Duration
	failChan chan struct{}
}

func NewTimeoutBlock(ctx context.Context, timeout time.Duration) (b *TimeoutBlock) {
	b = &TimeoutBlock{
		baseCtx: ctx,
		timeout: timeout,
		failChan: make(chan struct{}),
	}
	return
}

func (b *TimeoutBlock) Block() (err error) {
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

