package supervisor_test

import (
	"errors"
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
