package supervisor_test

import (
	"fmt"
	"github.com/akaspin/supervisor"
	"sync"
	"sync/atomic"
)

type dummyWatcher struct {
	in chan string
	wg sync.WaitGroup
	res []string
}

func newDummyWatcher(expect int, components ...*testingComponent) (w *dummyWatcher) {
	w = &dummyWatcher{
		in: make(chan string),
	}
	for _, comp := range components {
		comp.reportChan = w.in
	}
	w.wg.Add(expect)
	go func() {
		for m := range w.in {
			//println(m)
			w.res = append(w.res, m)
			w.wg.Done()
		}
	}()
	return
}


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
	reportChan chan string
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
	if c.reportChan != nil {
		c.reportChan<- c.name+"-open"
	}
	//println("o", c.name)
	return
}

func (c *testingComponent) Close() (err error) {
	atomic.AddInt32(&c.testingCounters.c, 1)
	err = c.errClose
	if c.reportChan != nil {
		c.reportChan<- c.name+"-close"
	}
	close(c.closedChan)
	//println("c", c.name)
	return
}

func (c *testingComponent) Wait() (err error) {
	atomic.AddInt32(&c.testingCounters.w, 1)
	err = c.errWait
	<-c.closedChan
	if c.reportChan != nil {
		c.reportChan<- c.name+"-wait"
	}
	//println("w", c.name)
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
