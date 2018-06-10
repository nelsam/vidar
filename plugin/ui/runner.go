// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package ui

// Runner is a type that can run functions on the UI goroutine.
type Runner interface {
	Wrapper

	// Queue queues the passed in function to run on the UI
	// goroutine, returning after adding it to the queue.
	Queue(func())

	// Run runs the passed in function on the UI goroutine,
	// returning after the function has run.
	Run(func())
}
