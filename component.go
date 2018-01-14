package supervisor

type Component interface {

	// Open blocks until component is opened
	Open() (err error)

	// Close blocks until component is closed
	Close() (err error)

	// Wait blocks until component is closed or error occurs
	Wait() (err error)
}
