package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

type testingCounters struct {
	openCn  int32
	closeCn int32
	waitCn  int32
}

type testingComponent struct {
	name     string
	counters testingCounters
	errOpen  error
	errClose error
	errWait  error
}

func (c *testingComponent) Open() (err error) {
	atomic.AddInt32(&c.counters.openCn, 1)
	err = c.errOpen
	return
}

func (c *testingComponent) Close() (err error) {
	atomic.AddInt32(&c.counters.closeCn, 1)
	err = c.errClose
	return
}

func (c *testingComponent) Wait() (err error) {
	atomic.AddInt32(&c.counters.waitCn, 1)
	err = c.errWait
	return
}

func assertStates(t *testing.T, expects []testingCounters, components []*testingComponent) {
	t.Helper()
	for i, c := range components {
		openOk := expects[i].openCn == -1 || atomic.LoadInt32(&c.counters.openCn) == expects[i].openCn
		closeOk := expects[i].closeCn == -1 || atomic.LoadInt32(&c.counters.closeCn) == expects[i].closeCn
		waitOk := expects[i].waitCn == -1 || atomic.LoadInt32(&c.counters.waitCn) == expects[i].waitCn
		if !openOk || !closeOk || !waitOk {
			t.Errorf("%d:%s: %v != %v", i, c.name, expects[i], c.counters)
		}
	}
}

func TestGroup1_Open(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			t.Run(`ok`, func(t *testing.T) {
				c1 := &testingComponent{}
				c2 := &testingComponent{}
				group := supervisor.NewGroup(context.Background(), c1, c2)
				assert.NoError(t, group.Open())

				assertStates(t,
					[]testingCounters{
						{1, 0, -1},
						{1, 0, -1},
					},
					[]*testingComponent{
						c1, c2,
					})
			})
			t.Run(`fail`, func(t *testing.T) {
				c1 := &testingComponent{}
				c2 := &testingComponent{
					errOpen: errors.New("c2"),
				}
				group := supervisor.NewGroup(context.Background(), c1, c2)
				assert.EqualError(t, group.Open(), "c2")

				time.Sleep(time.Millisecond * 50)
				assertStates(t,
					[]testingCounters{
						{1, 1, -1}, // first close should be called
						{1, 0, 0},  // second wait should not be called
					},
					[]*testingComponent{
						c1, c2,
					})
			})
		})
	}

}

func TestGroup1_Close(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			t.Run(`ok`, func(t *testing.T) {
				c1 := &testingComponent{}
				c2 := &testingComponent{}
				group := supervisor.NewGroup(context.Background(), c1, c2)

				assert.NoError(t, group.Open())
				assert.NoError(t, group.Close())
				assertStates(t,
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
					},
					[]*testingComponent{
						c1, c2,
					})
			})
			t.Run(`fail`, func(t *testing.T) {
				c1 := &testingComponent{}
				c2 := &testingComponent{
					errClose: errors.New("c2"),
				}
				group := supervisor.NewGroup(context.Background(), c1, c2)

				assert.NoError(t, group.Open())
				assert.EqualError(t, group.Close(), "c2")
				assertStates(t,
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
					},
					[]*testingComponent{
						c1, c2,
					})
			})
		})
	}
}

func TestGroup1_Wait(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			t.Run(`ok`, func(t *testing.T) {
				c1 := &testingComponent{}
				c2 := &testingComponent{}
				group := supervisor.NewGroup(context.Background(), c1, c2)

				assert.NoError(t, group.Open())
				assert.NoError(t, group.Close())
				assert.NoError(t, group.Wait())
				assertStates(t,
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
					},
					[]*testingComponent{
						c1, c2,
					})
			})
			t.Run(`fail after close`, func(t *testing.T) {
				c1 := &testingComponent{}
				c2 := &testingComponent{
					errWait: errors.New("c2"),
				}
				group := supervisor.NewGroup(context.Background(), c1, c2)

				assert.NoError(t, group.Open())
				assert.NoError(t, group.Close())
				assert.EqualError(t, group.Wait(), "c2")
				assertStates(t,
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
					},
					[]*testingComponent{
						c1, c2,
					})
			})
			t.Run(`fail without close`, func(t *testing.T) {
				c1 := &testingComponent{}
				c2 := &testingComponent{
					errWait: errors.New("c2"),
				}
				group := supervisor.NewGroup(context.Background(), c1, c2)

				assert.NoError(t, group.Open())
				assert.EqualError(t, group.Wait(), "c2")

				time.Sleep(time.Millisecond * 100)
				assertStates(t,
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
					},
					[]*testingComponent{
						c1, c2,
					})
			})
		})
	}
}
