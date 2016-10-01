package supervisor

import (
	"context"
	"sync"
)

// Group composes group of components. All component will be opened and
// closed together. Group "Wait" blocs until all components in chain are closed
// or error in at least in one components. On error whole group will be
// closed and "Wait" will return first error.
type Group struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup

	errMu   *sync.Mutex
	err        error

	trap *Trap

	components []Component
}

func NewGroup(ctx context.Context, components ...Component) (g *Group) {
	g = &Group{
		wg: &sync.WaitGroup{},
		errMu: &sync.Mutex{},
		components: components,
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	g.trap = newTrap(g.cancel)
	return
}

func (g *Group) Open() (err error) {
	//g.wg.Add(1)
	//go func() {
	//	<-g.ctx.Done()
	//	g.wg.Done()
	//}()

	for _, component := range g.components {
		if err = component.Open(); err != nil {
			g.trap.Catch(err)
			return
		}
		g.wg.Add(1)
		go g.supervise(component)
	}
	return
}

func (g *Group) Close() (err error) {
	g.cancel()
	return
}

func (g *Group) Wait() (err error) {
	g.wg.Wait()
	err = g.trap.Err()
	return
}

func (g *Group) supervise(component Component) {
	defer g.wg.Done()
	ctx, cancel := context.WithCancel(g.ctx)

	go func() {
		<-ctx.Done()
		if err := component.Close(); err != nil {
			g.trap.Catch(err)
		}
	}()

	if err := component.Wait(); err != nil {
		g.trap.Catch(err)
	}
	cancel()
}
