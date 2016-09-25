package supervisor

import (
	"io"
)

type Component interface {
	io.Closer

	Open() (err error)
	Wait() (err error)
}


