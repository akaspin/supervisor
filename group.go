package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
)

/*
Group supervises Components in parallel. All supervised components are open
and closed concurrently.

Open, Close and Wait methods may be called many times and will return equal
results. Group guarantees that Open, Close and Wait methods for all Components
will be called once.

Group collects and returns error from corresponding Component methods. If more
than one Components returns errors they will be wrapped in MultiError.
*/
type Group struct {
	*compositeBase
	components []Component
}

// NewGroup creates new Group. Provided context manages whole Group. Close
// Context is equivalent to call Group.Close().
func NewGroup(ctx context.Context, components ...Component) (g *Group) {
	g = &Group{
		compositeBase: newCompositeBase(ctx),
		components:    components,
	}
	return
}

// Open blocks until all components are opened. If Component Open() method
// returns error Group will close all ascendants.
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

// Close initialises shutdown for all Components.
func (c *Group) Close() (err error) {
	return c.compositeBase.Close()
}

// Wait blocks until all components in Group are exited. If one of Wait()
// method of one of Components is exited while Group is open Group closes all
// other Components in parallel.
func (c *Group) Wait() (err error) {
	return c.compositeBase.Wait()
}
