package supervisor_test

import (
	"context"
	"fmt"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestChain_Open(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run("ok_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(9, c1, c2, c3)

			chain := supervisor.NewChain(context.Background(), c1, c2, c3)

			assert.NoError(t, chain.Open())
			assert.NoError(t, chain.Open())

			go assert.NoError(t, chain.Close())
			assert.NoError(t, chain.Wait())
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)

			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open", "3-open",
				"3-close", "3-done",
				"2-close", "2-done",
				"1-close", "1-done",
			}, watcher.res)

			assert.NoError(t, chain.Open())
		})
		t.Run("fail"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", fmt.Errorf("2"), nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(4, c1, c2, c3)

			chain := supervisor.NewChain(context.Background(), c1, c2, c3)

			assert.EqualError(t, chain.Open(), "2")
			assert.EqualError(t, chain.Open(), "2")
			go assert.NoError(t, chain.Close())

			assert.NoError(t, chain.Wait())
			c1.assertCycle(t)
			c2.assertEvents(t, "open")
			c3.assertEvents(t)

			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open",
				"2-open",
				"1-close",
				"1-done",
			}, watcher.res)
		})
	}
}

func TestChain_Close(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run(`fail`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, fmt.Errorf("2"), nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(9, c1, c2, c3)

			chain := supervisor.NewChain(context.Background(), c1, c2, c3)

			assert.NoError(t, chain.Open())
			assert.EqualError(t, chain.Close(), "2")
			assert.NoError(t, chain.Wait())
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)
			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open", "3-open",
				"3-close", "3-done",
				"2-close", "2-done",
				"1-close", "1-done",
			}, watcher.res)

			assert.EqualError(t, chain.Close(), "2")
		})
	}
}

func TestChain_Wait(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run(`fail in the middle`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(8, c1, c2, c3)

			chain := supervisor.NewChain(context.Background(), c1, c2, c3)

			assert.NoError(t, chain.Open())
			close(c2.closedChan)
			assert.NoError(t, chain.Wait())
			c1.assertCycle(t)
			c2.assertEvents(t, "open", "done")
			c3.assertCycle(t)
			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open", "3-open",
				"2-done",
				"3-close", "3-done",
				"1-close", "1-done",
			}, watcher.res)

			assert.NoError(t, chain.Close())
		})
	}
}
