// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

// A BeforeExecutor is an Executor which has tasks to run on the
// Controller prior to running Exec.
type BeforeExecutor interface {
	BeforeExec(interface{})
}

type Elementer interface {
	Elements() []interface{}
}
