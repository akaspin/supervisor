package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestGroup_Open(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run(`ok-`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.NoError(t, group.Open())
			assert.NoError(t, group.Open()) // twice
			go group.Close()
			group.Wait()

			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)

			// open after close
			assert.NoError(t, group.Open())
			c1.assertCycle(t)
		})
		t.Run(`fail 2-`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", errors.New("2"), nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.EqualError(t, group.Open(), "2")
			assert.EqualError(t, group.Open(), "2") // twice

			group.Wait()
			c1.assertCycle(t)
			c2.assertEvents(t, "open")
			c3.assertCycle(t)

			// open after crash
			assert.EqualError(t, group.Open(), "2")
			c1.assertCycle(t)
			c2.assertEvents(t, "open")
		})
	}
}

func TestGroup_Close(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run(`ok`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			closedChan := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.EqualError(t, group.Close(), "not open")
			assert.EqualError(t, group.Wait(), "not open")

			assert.NoError(t, group.Open())

			go func() {
				group.Wait()
				close(closedChan)
			}()

			assert.NoError(t, group.Close())

			<-closedChan
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)

			// after
			assert.NoError(t, group.Close())
			c1.assertCycle(t)
		})
		t.Run(`fail`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			closedChan := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, errors.New("2"), nil)
			c3 := newTestingComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.EqualError(t, group.Close(), "not open")
			assert.EqualError(t, group.Wait(), "not open")

			assert.NoError(t, group.Open())

			go func() {
				group.Wait()
				close(closedChan)
			}()
			assert.EqualError(t, group.Close(), "2")

			<-closedChan
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)

			assert.EqualError(t, group.Close(), "2")
		})
	}
}

func TestGroup_Wait(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run(`ok`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			closedChan := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.EqualError(t, group.Close(), "not open")
			assert.EqualError(t, group.Wait(), "not open")

			assert.NoError(t, group.Open())

			// pre
			go func() {
				assert.NoError(t, group.Wait())
				c1.assertCycle(t)
				c2.assertCycle(t)
				c3.assertCycle(t)
				close(closedChan)
			}()

			group.Close()

			assert.NoError(t, group.Wait())
			assert.NoError(t, group.Wait())
		})
		t.Run(`fail`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			closedChan := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, errors.New("2"))
			c3 := newTestingComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.EqualError(t, group.Close(), "not open")
			assert.EqualError(t, group.Wait(), "not open")

			assert.NoError(t, group.Open())

			// pre
			go func() {
				assert.EqualError(t, group.Wait(), "2")
				c1.assertCycle(t)
				c2.assertCycle(t)
				c3.assertCycle(t)
				close(closedChan)
			}()

			group.Close()

			assert.EqualError(t, group.Wait(), "2")

			// second
			assert.EqualError(t, group.Wait(), "2")
		})
		t.Run(`self close`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.EqualError(t, group.Close(), "not open")
			assert.EqualError(t, group.Wait(), "not open")

			assert.NoError(t, group.Open())
			close(c2.closedChan)

			assert.NoError(t, group.Wait())
			c1.assertCycle(t)
			c2.assertEvents(t, "open", "done")
			c3.assertCycle(t)
		})
	}
}
