package supervisor

import (
	"context"
	"sync/atomic"
	"time"
)

// TimeoutBlocker
type TimeoutBlocker struct {
	baseCtx context.Context
	timeout time.Duration

	closed   int32
	failChan chan struct{}
}

func NewTimeoutBlocker(ctx context.Context, timeout time.Duration) (b *TimeoutBlocker) {
	b = &TimeoutBlocker{
		baseCtx:  ctx,
		timeout:  timeout,
		failChan: make(chan struct{}),
	}
	return
}

func (b *TimeoutBlocker) Open() (err error) {
	panic("implement me")
}

func (b *TimeoutBlocker) Close() (err error) {
	if !atomic.CompareAndSwapInt32(&b.closed, 0, 1) {
		return
	}
	go func() {
		time.Sleep(b.timeout)
		close(b.failChan)
	}()
	return
}

// Wait blocks until timeout is reached or TimeoutBlocker is closed.
// If timeout is reached Wait() will exit with error.
func (b *TimeoutBlocker) Wait() (err error) {
	select {
	case <-b.baseCtx.Done():
		return nil
	case <-b.failChan:
		err = context.DeadlineExceeded
	}
	return err
}
