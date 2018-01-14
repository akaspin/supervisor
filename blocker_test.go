package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

type dummyBlock struct {
	okChan   chan struct{}
	failChan chan struct{}
}

func (d *dummyBlock) Close() (err error) {
	return
}

func (d *dummyBlock) Wait() (err error) {
	select {
	case <-d.okChan:
	case <-d.failChan:
		err = errors.New("FAIL")
	}
	return
}

func newDummyBlock() (b *dummyBlock) {
	b = &dummyBlock{
		okChan:   make(chan struct{}),
		failChan: make(chan struct{}),
	}
	return
}

func TestCompositeBlock_Wait(t *testing.T) {
	t.Run("fail", func(t *testing.T) {
		b1 := newDummyBlock()
		b2 := newDummyBlock()
		blocker := supervisor.NewCompositeBlock(b1, b2)
		var done, fail int32

		// run wait first
		go func() {
			if err := blocker.Wait(); err != nil {
				atomic.AddInt32(&fail, 1)
			}
			atomic.AddInt32(&done, 1)
		}()

		go func() {
			blocker.Close()
			close(b2.failChan)
		}()

		// close both blockers
		time.Sleep(time.Millisecond * 100)

		assert.EqualError(t, blocker.Wait(), "FAIL")
		assert.Equal(t, int32(1), atomic.LoadInt32(&fail))
		assert.Equal(t, int32(1), atomic.LoadInt32(&done))
	})
}

func TestBlockTimeout_Wait(t *testing.T) {
	t.Run(`ok`, func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		blocker := supervisor.NewTimeoutBlock(ctx, time.Second)
		var done, fail int32

		// run wait first
		go func() {
			if err := blocker.Wait(); err != nil {
				atomic.AddInt32(&fail, 1)
			}
			atomic.AddInt32(&done, 1)
		}()

		go func() {
			blocker.Close()
		}()
		cancel()

		time.Sleep(time.Millisecond * 100)

		assert.NoError(t, blocker.Wait())
		assert.Equal(t, int32(0), atomic.LoadInt32(&fail))
		assert.Equal(t, int32(1), atomic.LoadInt32(&done))
	})
	t.Run(`timeout`, func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		blocker := supervisor.NewTimeoutBlock(ctx, time.Millisecond*10)
		var done, fail int32
		// run wait first
		go func() {
			if err := blocker.Wait(); err != nil {
				atomic.AddInt32(&fail, 1)
			}
			atomic.AddInt32(&done, 1)
		}()

		go func() {
			blocker.Close()
		}()

		time.Sleep(time.Millisecond * 100)
		assert.EqualError(t, blocker.Wait(), "context deadline exceeded")
		assert.Equal(t, int32(1), atomic.LoadInt32(&fail))
		assert.Equal(t, int32(1), atomic.LoadInt32(&done))
	})
}
