package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

const groupTestIterations = 10



func TestGroup_Open(t *testing.T) {
	for i := 0; i < groupTestIterations; i++ {
		t.Run(`ok-`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newDummyComponent("1", nil, nil, nil)
			c2 := newDummyComponent("2", nil, nil, nil)
			c3 := newDummyComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.NoError(t, group.Open())
			assert.NoError(t, group.Open()) // twice
			go group.Close()
			group.Wait()

			assert.NoError(t, assertComponents(
				[]*testingComponent{c1, c2, c3},
				[]testingCounters{
					{1, 1, -1},
					{1, 1, -1},
					{1, 1, -1},
				}))

			// open after close
			assert.NoError(t, group.Open())
			assert.NoError(t, assertComponents(
				[]*testingComponent{c1, c2, c3},
				[]testingCounters{
					{1, 1, -1},
					{1, 1, -1},
					{1, 1, -1},
				}))
		})
		t.Run(`fail 2-`+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			c1 := newDummyComponent("1", nil, nil, nil)
			c2 := newDummyComponent("2", errors.New("2"), nil, nil)
			c3 := newDummyComponent("3", nil, nil, nil)

			group := supervisor.NewGroup(context.Background(), c1, c2, c3)
			assert.EqualError(t, group.Open(), "2")
			assert.EqualError(t, group.Open(), "2") // twice

			group.Wait()
			assert.NoError(t, assertComponents(
				[]*testingComponent{c1, c2, c3},
				[]testingCounters{
					{1, 1, -1},
					{1, 0, -1}, // c2 should be not opened
					{1, 1, -1},
				}))

			// open after crash
			assert.EqualError(t, group.Open(), "2")
			assert.NoError(t, assertComponents(
				[]*testingComponent{c1, c2, c3},
				[]testingCounters{
					{1, 1, -1},
					{1, 0, -1}, // c2 should be not opened
					{1, 1, -1},
				}))
		})
	}
}

func TestGroup_Close(t *testing.T) {
	for i := 0; i < groupTestIterations; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			t.Run(`ok`, func(t *testing.T) {
				closedChan := make(chan struct{})
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", nil, nil, nil)
				c3 := newDummyComponent("3", nil, nil, nil)

				group := supervisor.NewGroup(context.Background(), c1, c2, c3)
				assert.EqualError(t, group.Close(), "not open")
				assert.EqualError(t, group.Wait(), "not open")

				assert.NoError(t, group.Open())

				go func() {
					group.Wait()
					close(closedChan)
				}()

				assert.NoError(t, group.Close())

				<-closedChan
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
						{1, 1, -1},
					}))

				// after
				assert.NoError(t, group.Close())
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, -1},
						{1, 1, -1},
						{1, 1, -1},
					}))
			})
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				t.Run(`fail`, func(t *testing.T) {
					closedChan := make(chan struct{})
					c1 := newDummyComponent("1", nil, nil, nil)
					c2 := newDummyComponent("2", nil, errors.New("2"), nil)
					c3 := newDummyComponent("3", nil, nil, nil)

					group := supervisor.NewGroup(context.Background(), c1, c2, c3)
					assert.EqualError(t, group.Close(), "not open")
					assert.EqualError(t, group.Wait(), "not open")

					assert.NoError(t, group.Open())

					go func() {
						group.Wait()
						close(closedChan)
					}()
					assert.EqualError(t, group.Close(), "2")

					<-closedChan
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, -1},
							{1, 1, -1},
							{1, 1, -1},
						}))

					assert.EqualError(t, group.Close(), "2")
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, -1},
							{1, 1, -1},
							{1, 1, -1},
						}))
				})
			})
		})
	}
}

func TestGroup_Wait(t *testing.T) {
	for i := 0; i < groupTestIterations; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			t.Run(`ok`, func(t *testing.T) {
				closedChan := make(chan struct{})
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", nil, nil, nil)
				c3 := newDummyComponent("3", nil, nil, nil)

				group := supervisor.NewGroup(context.Background(), c1, c2, c3)
				assert.EqualError(t, group.Close(), "not open")
				assert.EqualError(t, group.Wait(), "not open")

				assert.NoError(t, group.Open())

				// pre
				go func() {
					assert.NoError(t, group.Wait())
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, 1},
							{1, 1, 1},
							{1, 1, 1},
						}))
					close(closedChan)
				}()

				group.Close()

				assert.NoError(t, group.Wait())
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))

				// second
				assert.NoError(t, group.Wait())
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))
			})
			t.Run(`fail`, func(t *testing.T) {
				closedChan := make(chan struct{})
				c1 := newDummyComponent("1", nil, nil, nil)
				c2 := newDummyComponent("2", nil, nil, errors.New("2"))
				c3 := newDummyComponent("3", nil, nil, nil)

				group := supervisor.NewGroup(context.Background(), c1, c2, c3)
				assert.EqualError(t, group.Close(), "not open")
				assert.EqualError(t, group.Wait(), "not open")

				assert.NoError(t, group.Open())

				// pre
				go func() {
					assert.EqualError(t, group.Wait(), "2")
					assert.NoError(t, assertComponents(
						[]*testingComponent{c1, c2, c3},
						[]testingCounters{
							{1, 1, 1},
							{1, 1, 1},
							{1, 1, 1},
						}))
					close(closedChan)
				}()

				group.Close()

				assert.EqualError(t, group.Wait(), "2")
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))

				// second
				assert.EqualError(t, group.Wait(), "2")
				assert.NoError(t, assertComponents(
					[]*testingComponent{c1, c2, c3},
					[]testingCounters{
						{1, 1, 1},
						{1, 1, 1},
						{1, 1, 1},
					}))
			})
		})
	}
}
