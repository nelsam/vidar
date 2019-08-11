// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

// A BeforeExecutor is an Executor which has tasks to run
// prior to running Exec.
type BeforeExecutor interface {
	BeforeExec(interface{})
}

// Elementer is a type which contains elements of its own.
type Elementer interface {
	Elements() []interface{}
}
