package supervisor

import (
	"sync"
	"context"
)


type Trap struct {
	cancel context.CancelFunc
	err error
	errMu *sync.Mutex
}

func NewTrap(cancel context.CancelFunc) (t *Trap) {
	t = &Trap{
		cancel: cancel,
		errMu: &sync.Mutex{},
	}
	return
}

func (t *Trap) Err() (err error) {
	t.errMu.Lock()
	defer t.errMu.Unlock()
	err = t.err
	return
}

func (t *Trap) Catch(err error) {
	t.errMu.Lock()
	defer t.errMu.Unlock()
	if t.err == nil {
		t.err = err
	}
	t.cancel()
}
