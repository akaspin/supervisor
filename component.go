package supervisor

// Component is basic building block to build supervisor trees.
// It recommends to implement all methods to can be called multiple
// times and return equal results.
type Component interface {
	// Open runs Component initialisation and should blocks until
	// Component is initialised. Open() method should be called before
	// Wait() or Close().
	Open() (err error)
	// Close initialises Component shutdown.
	// Non blocking behaviour is recommended.
	Close() (err error)
	// Wait should blocks until Component shutdown.
	Wait() (err error)
}
