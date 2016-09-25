package supervisor

import (
	"sync"
	"context"
)

type trap struct {
	cancel context.CancelFunc
	err error
	errMu *sync.Mutex
}

func newTrap(cancel context.CancelFunc) (t *trap) {
	t = &trap{
		cancel: cancel,
		errMu: &sync.Mutex{},
	}
	return
}

func (t *trap) Err() (err error) {
	t.errMu.Lock()
	defer t.errMu.Unlock()
	err = t.err
	return
}

func (t *trap) trapErr(err error) {
	t.errMu.Lock()
	defer t.errMu.Unlock()
	if t.err == nil {
		t.err = err
	}
	t.cancel()
}
