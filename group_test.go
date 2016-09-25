package supervisor_test

import (
	"testing"
	"sync"
	"context"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"errors"
	"sync/atomic"
)

type crashable struct {
	ctx context.Context
	cancel context.CancelFunc

	openC *int64
	doneC *int64
	waitC *int64
	wg *sync.WaitGroup
	err error
}

func newCrashable(openC, doneC, waitC *int64) (c *crashable) {
	c = &crashable{
		openC: openC,
		doneC: doneC,
		waitC: waitC,
		wg: &sync.WaitGroup{},
	}
	c.ctx, c.cancel = context.WithCancel(context.TODO())
	return
}

func (c *crashable) Open() (err error) {
	c.wg.Add(1)
	go func() {
		<-c.ctx.Done()
		c.wg.Done()
		atomic.AddInt64(c.doneC, 1)
	}()
	atomic.AddInt64(c.openC, 1)
	return
}

func (c *crashable) Close() (err error) {
	c.cancel()
	return
}

func (c *crashable) Wait() (err error) {
	c.wg.Wait()
	atomic.AddInt64(c.waitC, 1)
	err = c.err
	return
}

func (c *crashable) Crash(err error) {
	c.err = err
	c.Close()
}

func TestGroup_Empty(t *testing.T) {
	g := supervisor.NewGroup(context.TODO())
	g.Open()
	//g.Close()
	g.Wait()
}

func TestGroup_Regular(t *testing.T) {
	var openC, doneC, waitC int64
	g := supervisor.NewGroup(
		context.TODO(),
		newCrashable(&openC, &doneC, &waitC),
		newCrashable(&openC, &doneC, &waitC),
		newCrashable(&openC, &doneC, &waitC),
	)
	g.Open()
	g.Close()
	err := g.Wait()
	assert.NoError(t, err)
	assert.Equal(t, []int64{3, 3, 3}, []int64{openC, doneC, waitC})
}

func TestGroup_Crash(t *testing.T) {
	var openC, doneC, waitC int64
	messy := newCrashable(&openC, &doneC, &waitC)
	g := supervisor.NewGroup(
		context.TODO(),
		newCrashable(&openC, &doneC, &waitC),
		newCrashable(&openC, &doneC, &waitC),
		messy,
	)
	g.Open()
	messy.Crash(errors.New("err"))
	err := g.Wait()
	assert.Error(t, err)
	assert.Equal(t, "err", err.Error())
	assert.Equal(t, []int64{3, 3, 3}, []int64{openC, doneC, waitC})
}
