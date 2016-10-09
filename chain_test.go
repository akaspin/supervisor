package supervisor_test

import (
	"testing"
	"github.com/akaspin/supervisor"
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"errors"
)

type chainableSig struct {
	index int
	op string
}

type chainable struct {
	ctx context.Context
	cancel context.CancelFunc
	wg *sync.WaitGroup

	index int

	ch chan chainableSig
	err error
}

func newChainable(index int, rCh chan chainableSig) (c *chainable) {
	c = &chainable{
		index: index,
		ch: rCh,
		wg: &sync.WaitGroup{},
	}
	c.ctx, c.cancel = context.WithCancel(context.TODO())
	return
}

func (c *chainable) Open() (err error) {
	c.wg.Add(1)
	go func() {
		<-c.ctx.Done()
		c.wg.Done()
		c.ch<-chainableSig{c.index, "done"}
	}()
	c.ch<-chainableSig{c.index, "open"}
	return
}

func (c *chainable) Close() (err error) {
	c.cancel()
	return
}

func (c *chainable) Wait() (err error) {
	c.wg.Wait()
	c.ch<-chainableSig{c.index, "wait"}
	err = c.err
	return
}

func (c *chainable) Crash(err error) {
	c.err = err
	c.Close()
}

func TestChain_Empty(t *testing.T) {
	c := supervisor.NewChain(context.TODO())
	c.Open()
	//c.Close()
	c.Wait()
}

func TestChain_OK(t *testing.T) {
	resCh := make(chan chainableSig)
	var res []chainableSig
	resWg := &sync.WaitGroup{}
	resWg.Add(1)
	go func() {
		for i:=0;i<9;i++ {
			res = append(res, <-resCh)
		}
		resWg.Done()
	}()


	c := supervisor.NewChain(
		context.TODO(),
		newChainable(1, resCh),
		newChainable(2, resCh),
		newChainable(3, resCh),
	)
	c.Open()
	c.Close()
	err := c.Wait()
	resWg.Wait()

	assert.NoError(t, err)
	assert.Equal(t, res, []chainableSig{
		{index:1, op:"open"},
		{index:2, op:"open"},
		{index:3, op:"open"},
		{index:3, op:"done"},
		{index:3, op:"wait"},
		{index:2, op:"done"},
		{index:2, op:"wait"},
		{index:1, op:"done"},
		{index:1, op:"wait"}})
}

func TestChain_Crash(t *testing.T) {
	resCh := make(chan chainableSig)
	var res []chainableSig
	resWg := &sync.WaitGroup{}
	resWg.Add(1)
	go func() {
		for i:=0;i<9;i++ {
			res = append(res, <-resCh)
		}
		resWg.Done()
	}()

	messy := newChainable(1, resCh)
	g := supervisor.NewChain(
		context.TODO(),
		newChainable(2, resCh),
		newChainable(3, resCh),
		messy,
	)
	g.Open()
	messy.Crash(errors.New("err"))
	err := g.Wait()
	resWg.Wait()

	assert.Error(t, err)
	assert.Equal(t, "err", err.Error())
	assert.Equal(t, res, []chainableSig{
		{index:2, op:"open"},
		{index:3, op:"open"},
		{index:1, op:"open"},
		{index:1, op:"done"},
		{index:1, op:"wait"},
		{index:3, op:"done"},
		{index:3, op:"wait"},
		{index:2, op:"done"},
		{index:2, op:"wait"}})
}
