package supervisor_test

import (
	"testing"
	"github.com/akaspin/supervisor"
	"context"
	"github.com/stretchr/testify/assert"
	"errors"
)

func TestChain_Empty(t *testing.T) {
	c := supervisor.NewChain(context.TODO())
	c.Open()
	c.Close()
	c.Wait()
}

func TestChain_OK(t *testing.T) {
	var openC, doneC, waitC int64
	c := supervisor.NewChain(
		context.TODO(),
		newCrashable(&openC, &doneC, &waitC),
		newCrashable(&openC, &doneC, &waitC),
		newCrashable(&openC, &doneC, &waitC),
	)
	c.Open()
	c.Close()
	err := c.Wait()
	assert.NoError(t, err)
	assert.Equal(t, []int64{3, 3, 3}, []int64{openC, doneC, waitC})
}

func TestChain_Crash(t *testing.T) {
	var openC, doneC, waitC int64
	messy := newCrashable(&openC, &doneC, &waitC)
	g := supervisor.NewChain(
		context.TODO(),
		newCrashable(&openC, &doneC, &waitC),
		newCrashable(&openC, &doneC, &waitC),
		messy,
	)
	g.Open()
	messy.Crash(errors.New("err"))
	err := g.Wait()
	assert.Error(t, err)
	assert.Equal(t, "err", err.Error())
	assert.Equal(t, []int64{3, 3, 3}, []int64{openC, doneC, waitC})
}
