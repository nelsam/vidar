// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import "github.com/nelsam/gxui"

// Driver wraps gxui.Driver as a ui.Runner.
type Driver struct {
	driver gxui.Driver
}

// Parent returns the gxui.Driver.
func (r *Runner) Parent() interface{} {
	return r.driver
}

// Queue queues f to run on the UI goroutine, returning once it is
// enqueued.
func (r *Runner) Queue(f func()) {
	r.driver.Call(f)
}

// Run runs f on the UI goroutine, returning once it has run.
func (r *Runner) Run(f func()) {
	r.driver.CallSync(f)
}
