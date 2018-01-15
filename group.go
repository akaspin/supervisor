package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
)

// Group manages components in parallel manner. Empty Group closes
// immediately with error.
type Group struct {
	ctx    context.Context
	cancel context.CancelFunc

	components []Component

	opened uint32

	openE componentErr

	closeWg sync.WaitGroup
	closeE  componentErr

	waitWg sync.WaitGroup
	waitE  componentErr
}

// NewGroup creates Group. Provided context manages whole group.
func NewGroup(ctx context.Context, components ...Component) (g *Group) {
	g = &Group{
		components: components,
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	return
}

// Open opens all components together and blocks until they are opened.
// If at least one of components returns error Group closes all components
// and returns all Open errors. Open may be called many times and guarantees
// that Open of each component will be called only once. Regardless of number
// of calls Open always returns same result.
func (g *Group) Open() (err error) {
	if !atomic.CompareAndSwapUint32(&g.opened, 0, 1) {
		return g.openE.get()
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(g.components))
	for _, component := range g.components {
		go func(component Component) {
			defer wg.Done()
			if openErr := component.Open(); openErr != nil {
				g.openE.set(openErr)
				g.cancel()
				return
			}
			// close watchdog
			g.closeWg.Add(1)
			go func() {
				defer g.closeWg.Done()
				<-g.ctx.Done()
				if closeErr := component.Close(); closeErr != nil {
					g.closeE.set(closeErr)
				}
			}()
			// wait watchdog
			g.waitWg.Add(1)
			go func() {
				defer g.waitWg.Done()
				if waitErr := component.Wait(); waitErr != nil {
					g.waitE.set(waitErr)
				}
				g.cancel()
			}()
		}(component)
	}
	// wait
	wg.Wait()
	err = g.openE.error
	return
}

// Close closes all components and returns Close errors if any.
// Close may be called multiple times and guarantees that each component will
// be closed only once. Regardless of number of calls Close will return same
// result. If Group is not opened yet Close will return `ErrNotOpened`.
func (g *Group) Close() (err error) {
	if atomic.LoadUint32(&g.opened) == 0 {
		return ErrNotOpened
	}
	g.cancel()
	g.closeWg.Wait()
	err = g.closeE.get()
	return err
}

// Wait waits for all components. If one of Component.Wait returns Wait
// closes all components. Wait may be called multiple times and guarantees
// that Wait for each component will be called only once. Regardless of number
// of calls Wait will return same result. If Group is not opened yet Wait
// will return `ErrNotOpened`.
func (g *Group) Wait() (err error) {
	if atomic.LoadUint32(&g.opened) == 0 {
		return ErrNotOpened
	}
	g.waitWg.Wait()
	err = g.waitE.get()
	return
}
