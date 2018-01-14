package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWrapped_OK(t *testing.T) {
	fn := func() error {
		return nil
	}
	w := supervisor.NewWrapped(context.TODO(), fn)
	w.Open()
	w.Close()
	err := w.Wait()
	assert.NoError(t, err)
}

func TestWrapped_Err(t *testing.T) {
	fn := func() error {
		return errors.New("err")
	}
	w := supervisor.NewWrapped(context.TODO(), fn)
	w.Open()
	w.Close()
	err := w.Wait()
	assert.Error(t, err)
}
