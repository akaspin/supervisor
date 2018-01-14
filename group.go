package supervisor

import (
	"context"
	"sync"
)

// Group manages components in parallel
type Group struct {
	ctx    context.Context
	cancel context.CancelFunc

	components []Component

	closeWg    sync.WaitGroup
	closeErrMu sync.Mutex
	closeErr   error

	waitWg    sync.WaitGroup
	waitErrMu sync.Mutex
	waitErr   error
}

// NewGroup creates Group. Provided context manages whole group.
func NewGroup(ctx context.Context, components ...Component) (g *Group) {
	g = &Group{
		components: components,
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	return
}

// Open starts all components together and blocks until they are opened.
// If at least one of components returns error Group closes all components
// and returns all Open errors.
func (g *Group) Open() (err error) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(g.components))
	for _, component := range g.components {
		go func(component Component) {
			defer wg.Done()
			if openErr := component.Open(); openErr != nil {
				mu.Lock()
				err = AppendError(err, openErr)
				mu.Unlock()
				g.cancel()
				return
			}
			// close watchdog
			g.closeWg.Add(1)
			go func() {
				defer g.closeWg.Done()
				<-g.ctx.Done()
				if closeErr := component.Close(); closeErr != nil {
					g.closeErrMu.Lock()
					g.closeErr = AppendError(g.closeErr, closeErr)
					g.closeErrMu.Unlock()
				}
			}()
			// wait watchdog
			g.waitWg.Add(1)
			go func() {
				defer g.waitWg.Done()
				if waitErr := component.Wait(); waitErr != nil {
					g.waitErrMu.Lock()
					g.waitErr = AppendError(g.waitErr, waitErr)
					g.waitErrMu.Unlock()
					g.cancel()
				}
			}()
		}(component)
	}
	// wait
	wg.Wait()
	return
}

// Close closes all components and returns Close errors
func (g *Group) Close() (err error) {
	g.cancel()
	g.closeWg.Wait()
	err = g.closeErr
	return
}

// Wait waits for all components. If one of Component.Wait returns error Wait
// closes all components.
func (g *Group) Wait() (err error) {
	g.waitWg.Wait()
	err = g.waitErr
	return
}
