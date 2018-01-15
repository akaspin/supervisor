package supervisor_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync/atomic"
	"testing"
)

const groupTestIterations = 10

type testingCounters struct {
	o int32
	c int32
	w int32
}

type testingComponent struct {
	name string
	testingCounters
	errOpen  error
	errClose error
	errWait  error

	closedChan chan struct{}
}

func newDummyComponent(name string, errOpen, errClose, errWait error) (c *testingComponent) {
	c = &testingComponent{
		name:       name,
		errOpen:    errOpen,
		errClose:   errClose,
		errWait:    errWait,
		closedChan: make(chan struct{}),
	}
	return
}

func (c *testingComponent) Open() (err error) {
	atomic.AddInt32(&c.testingCounters.o, 1)
	err = c.errOpen
	return
}

func (c *testingComponent) Close() (err error) {
	atomic.AddInt32(&c.testingCounters.c, 1)
	err = c.errClose
	close(c.closedChan)
	return
}

func (c *testingComponent) Wait() (err error) {
	atomic.AddInt32(&c.testingCounters.w, 1)
	err = c.errWait
	<-c.closedChan
	return
}

func (c *testingComponent) assertCounters(oc, cc, wc int32) (err error) {
	openC := atomic.LoadInt32(&c.testingCounters.o)
	closeC := atomic.LoadInt32(&c.testingCounters.c)
	waitC := atomic.LoadInt32(&c.testingCounters.w)
	oOk := oc == -1 || openC == oc
	cOk := cc == -1 || closeC == cc
	wOk := wc == -1 || waitC == wc
	if !(oOk && cOk && wOk) {
		err = fmt.Errorf("%s: open(%t)=%d:%d close(%t)=%d:%d wait(%t)=%d:%d", c.name,
			oOk, oc, openC,
			cOk, cc, closeC,
			wOk, wc, waitC)
	}
	return
}

func assertComponents(components []*testingComponent, counters []testingCounters) (err error) {
	for i, c := range components {
		if err1 := c.assertCounters(counters[i].o, counters[i].c, counters[i].w); err != nil {
			err = supervisor.AppendError(err, err1)
		}
	}
	return
}

func TestGroup_Open(t *testing.T) {
	for i := 0; i < groupTestIterations; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			t.Run(`ok`, func(t *testing.T) {
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", nil, nil, nil)
				c3 := newDummyComponent("3", nil, nil, nil)

				group := supervisor.NewGroup(context.Background(), c1, c2, c3)
				assert.NoError(t, group.Open())
				assert.NoError(t, group.Open()) // twice
				go group.Close()
				group.Wait()

				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
						{1, 1, -1},
					}))

				// open after close
				assert.NoError(t, group.Open())
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
						{1, 1, -1},
					}))
			})
			t.Run(`fail 2`, func(t *testing.T) {
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", errors.New("2"), nil, nil)
				c3 := newDummyComponent("3", nil, nil, nil)

				group := supervisor.NewGroup(context.Background(), c1, c2, c3)
				assert.EqualError(t, group.Open(), "2")
				assert.EqualError(t, group.Open(), "2") // twice

				group.Wait()
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 0, -1}, // c2 should be not opened
						{1, 1, -1},
					}))

				// open after crash
				assert.EqualError(t, group.Open(), "2")
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 0, -1}, // c2 should be not opened
						{1, 1, -1},
					}))
			})
		})
	}
}

func TestGroup_Close(t *testing.T) {
	for i := 0; i < groupTestIterations; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			t.Run(`ok`, func(t *testing.T) {
				closedChan := make(chan struct{})
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", nil, nil, nil)
				c3 := newDummyComponent("3", nil, nil, nil)

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
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
						{1, 1, -1},
					}))

				// after
				assert.NoError(t, group.Close())
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
						{1, 1, -1},
					}))
			})
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				t.Run(`fail`, func(t *testing.T) {
					closedChan := make(chan struct{})
					c1 := newDummyComponent("1", nil, nil, nil)
					c2 := newDummyComponent("2", nil, errors.New("2"), nil)
					c3 := newDummyComponent("3", nil, nil, nil)

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
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, -1},
							{1, 1, -1},
							{1, 1, -1},
						}))

					assert.EqualError(t, group.Close(), "2")
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, -1},
							{1, 1, -1},
							{1, 1, -1},
						}))
				})
			})
		})
	}
}

func TestGroup_Wait(t *testing.T) {
	for i := 0; i < groupTestIterations; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			t.Run(`ok`, func(t *testing.T) {
				closedChan := make(chan struct{})
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", nil, nil, nil)
				c3 := newDummyComponent("3", nil, nil, nil)

				group := supervisor.NewGroup(context.Background(), c1, c2, c3)
				assert.EqualError(t, group.Close(), "not open")
				assert.EqualError(t, group.Wait(), "not open")

				assert.NoError(t, group.Open())

				// pre
				go func() {
					assert.NoError(t, group.Wait())
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, 1},
							{1, 1, 1},
							{1, 1, 1},
						}))
					close(closedChan)
				}()

				group.Close()

				assert.NoError(t, group.Wait())
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))

				// second
				assert.NoError(t, group.Wait())
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))
			})
			t.Run(`fail`, func(t *testing.T) {
				closedChan := make(chan struct{})
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", nil, nil, errors.New("2"))
				c3 := newDummyComponent("3", nil, nil, nil)

				group := supervisor.NewGroup(context.Background(), c1, c2, c3)
				assert.EqualError(t, group.Close(), "not open")
				assert.EqualError(t, group.Wait(), "not open")

				assert.NoError(t, group.Open())

				// pre
				go func() {
					assert.EqualError(t, group.Wait(), "2")
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, 1},
							{1, 1, 1},
							{1, 1, 1},
						}))
					close(closedChan)
				}()

				group.Close()

				assert.EqualError(t, group.Wait(), "2")
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))

				// second
				assert.EqualError(t, group.Wait(), "2")
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))
			})
		})
	}
}
