package supervisor

import (
	"context"
	"sync"
)


type link struct {
	ascendant *link
	trap *trap
	wg *sync.WaitGroup

	component Component
	ctx context.Context
	cancel context.CancelFunc
}

func (l *link) supervise() {
	ctx, cancel := context.WithCancel(context.Background())

	// set upstream watchdog
	go func() {
		defer l.wg.Done()
		LOOP:
		for {
			select {
			case <-ctx.Done():
				// wait reached
				if l.ascendant != nil {
					l.ascendant.cancel()
				}
				break LOOP
			case <-l.ctx.Done():
				// external close
				if err := l.component.Close(); err != nil {
					l.trap.trapErr(err)
				}
				continue LOOP
			}
		}
	}()

	if err := l.component.Wait(); err != nil {
		l.trap.trapErr(err)
	}
	cancel()
}

// Chain provides "chain of responsibility". Ascendants always
// opens before descendants and closes after
type Chain struct {
	ctx context.Context
	cancel context.CancelFunc
	wg *sync.WaitGroup

	components []Component
	trap *trap
}

func NewChain(ctx context.Context, components ...Component) (c *Chain) {
	c = &Chain{
		wg: &sync.WaitGroup{},
		components: components,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.trap = newTrap(c.cancel)
	return
}

func (c *Chain) Open() (err error) {
	c.wg.Add(1)

	background := context.Background()
	var ascendant *link
	for _, component := range c.components {
		if err = component.Open(); err != nil {
			c.Close()
			return
		}
		c.wg.Add(1)
		l := &link{
			ascendant: ascendant,
			trap: c.trap,
			wg: c.wg,
			component: component,
		}
		l.ctx, l.cancel = context.WithCancel(background)
		background = l.ctx
		ascendant = l
		go l.supervise()
	}

	go func() {
		<-c.ctx.Done()
		if ascendant != nil {
			ascendant.cancel()
		}
		c.wg.Done()
	}()
	return
}

func (c *Chain) Close() (err error) {
	c.cancel()
	return
}

func (c *Chain) Wait() (err error) {
	c.wg.Wait()
	err = c.trap.Err()
	return
}
