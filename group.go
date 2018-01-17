package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
)

// Group manages components in parallel manner. Empty Group closes
// immediately with error.
type Group struct {
	*compositeBase
	components []Component
}

// NewGroup creates Group. Provided context manages whole group.
func NewGroup(ctx context.Context, components ...Component) (g *Group) {
	g = &Group{
		compositeBase: newCompositeBase(ctx),
		components:    components,
	}
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
			var waitExited uint32
			go func() {
				defer g.closeWg.Done()
				<-g.ctx.Done()
				if atomic.CompareAndSwapUint32(&waitExited, 0, 1) {
					if closeErr := component.Close(); closeErr != nil {
						g.closeE.set(closeErr)
					}
				}
			}()
			// wait watchdog
			g.waitWg.Add(1)
			go func() {
				defer g.waitWg.Done()
				if waitErr := component.Wait(); waitErr != nil {
					g.waitE.set(waitErr)
				}
				atomic.CompareAndSwapUint32(&waitExited, 0, 1)
				g.cancel()
			}()
		}(component)
	}
	wg.Wait()
	err = g.openE.error
	return
}
