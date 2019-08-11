// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import "github.com/nelsam/gxui"

// Runner wraps gxui.Driver as a ui.Runner.
type Runner struct {
	driver gxui.Driver
}

// Raw returns the raw gxui.Driver.  This is temporary, similar to
// Creator.Raw().
func (r Runner) Raw() gxui.Driver {
	return r.driver
}

// Parent returns the gxui.Driver.
func (r Runner) Parent() interface{} {
	return r.driver
}

// Queue queues f to run on the UI goroutine, returning once it is
// enqueued.
func (r Runner) Enqueue(f func()) {
	r.driver.Call(f)
}

// Run runs f on the UI goroutine, returning once it has finished.
func (r Runner) Run(f func()) {
	r.driver.CallSync(f)
}
