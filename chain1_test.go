package supervisor_test

import (
	"context"
	"fmt"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestChain1_Open(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Run("ok_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newDummyComponent("1", nil, nil, nil)
			c2 := newDummyComponent("2", nil, nil, nil)
			c3 := newDummyComponent("3", nil, nil, nil)
			watcher := newDummyWatcher(9, c1, c2, c3)

			chain := supervisor.NewChain1(context.Background(), c1, c2, c3)

			assert.NoError(t, chain.Open())
			assert.NoError(t, chain.Open())

			assert.NoError(t, chain.Close())
			assert.NoError(t, chain.Wait())

			assert.NoError(t, assertComponents(
				[]*testingComponent{c1, c2, c3},
				[]testingCounters{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				}))
			watcher.wg.Wait()
			assert.Equal(t, []string{"" +
				"1-open", "2-open", "3-open",
				"3-close", "3-wait",
				"2-close", "2-wait",
				"1-close", "1-wait",
			}, watcher.res)

			assert.NoError(t, chain.Open())
			assert.NoError(t, assertComponents(
				[]*testingComponent{c1, c2, c3},
				[]testingCounters{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				}))
		})
		t.Run("fail"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newDummyComponent("1", nil, nil, nil)
			c2 := newDummyComponent("2", fmt.Errorf("2"), nil, nil)
			c3 := newDummyComponent("3", nil, nil, nil)
			watcher := newDummyWatcher(4, c1, c2, c3)

			chain := supervisor.NewChain1(context.Background(), c1, c2, c3)

			assert.NoError(t, chain.Open())
			assert.NoError(t, chain.Open())
			go assert.NoError(t, chain.Close())

			assert.NoError(t, chain.Wait())
			assert.NoError(t, assertComponents(
				[]*testingComponent{c1, c2, c3},
				[]testingCounters{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				}))
			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open",
				"2-open",
				"1-close",
				"1-wait",
			}, watcher.res)
		})
	}
}
