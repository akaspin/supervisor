package supervisor_test

import (
	"context"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBlockTimeout_Wait(t *testing.T) {
	t.Run(`ok`, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		blocker := supervisor.NewTimeoutBlock(ctx, time.Second)

		go func() {
			blocker.Block()
		}()
		cancel()
		assert.NoError(t, blocker.Wait())
	})
	t.Run(`timeout`, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		blocker := supervisor.NewTimeoutBlock(ctx, time.Millisecond*10)

		go func() {
			blocker.Block()
		}()
		assert.EqualError(t, blocker.Wait(), "context deadline exceeded")

	})
}
