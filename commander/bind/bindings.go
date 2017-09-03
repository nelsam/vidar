// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package bind contains some of the types that the commander
// package uses to identify types to bind to user input or other
// events.  It is kept in a separate package so that functions
// and methods can use these types in their signature without
// importing the commander package.  The intent is to allow
// plugins to continue to work without the need to rebuild them
// every time the commander package changes.
//
// For more information, see vidar's plugin package documentation.
package bind

// A Bindable is the base type for all types that can be bound to
// events.
type Bindable interface {
	// Name returns the name of the Bindable.  It must be unique.
	Name() string
}

// Op is a type that can execute an operation.
type Op interface {
	Bindable

	// Exec executes the op.  It will be called once for each
	// element in the editor, including all other Bindables, and
	// should return a status indicating its state after executing
	// against elem.
	//
	// A status that contains the Stop bit will mean that Exec
	// will not be called again.
	Exec(elem interface{}) Status
}

// HookedOp is an Op that allows Bindables to bind to events that
// are created by it.
type HookedOp interface {
	Op

	// Bind will be called for each Bindable that wants to bind to
	// events on the HookedOp.  A new HookedOp should be returned
	// with the Bindable bound to it, so that the original HookedOp
	// may continue to execute without the Bindable.
	Bind(Bindable) (HookedOp, error)
}

// A MultiOp is a type that makes it easier to execute using
// multiple types in the editor hierarchy.  The MultiOp can
// gather types that it needs before Exec is called.
type MultiOp interface {
	Bindable

	// Reset must be implemented to reset the MultiExecutor's state
	Reset()

	// Store is called in the same way that Op.Exec is called;
	// once for each type in vidar's state.  Values may be
	// stored by the MultiOp to prepare for the final Exec
	// call.
	Store(interface{}) Status

	// Exec is called after Store has finished.  The return value
	// is a simple error because there are only two states it can
	// be in after running - success or failure.  There is no
	// "waiting" state for a MultiOp's Exec method.
	Exec() error
}

// HookedMultiOp is to MultiOp as HookedOp is to Op.
type HookedMultiOp interface {
	MultiOp

	// Bind is equivalent to HookedOp.Bind, but with different
	// return values.
	Bind(Bindable) (HookedMultiOp, error)
}

// An OpHook is a binding that is bound to an Op.  It is up to the
// OpHook to ensure that it implements the interface required by the
// Op it is being bound to..
type OpHook interface {
	Bindable

	// OpName returns the name of the binding that this hook
	// wants to bind to.
	OpName() string
}

// A Command is a binding that happens due to the user explicitly
// requesting an event, usually via a key binding.
type Command interface {
	Bindable

	// Menu returns the name of the menu that the command should be
	// displayed under.  All user-executed Bindings *must* have an
	// entry in the menu to make them more discoverable.
	Menu() string
}

type Status int

const (
	// Waiting means that the Executor is still waiting for types
	// that it *must* execute against.
	Waiting Status = 1 << iota

	// Executed means that the Executor has successfully executed.
	Executed

	// Errored means that a required process in the Executor has
	// failed.
	Errored

	// Stop means that the Executor should stop executing.  It does
	// not mean that there was a success or failure, but that there
	// is nothing more to do.
	Stop

	// Executing means that the Executor has performed all required
	// operations, but there may be extra operations to perform.
	// This is most useful when an Executor *may* have more than
	// one type to execute against, but it only *needs* to execute
	// once.
	Executing = Executed | Waiting

	// Done means that the Executor has successfully performed
	// everything it wanted to perform and is finished.
	Done = Executed | Stop

	// Failed means that the Executor has errored and cannot perform
	// any additional tasks.
	Failed = Errored | Stop
)
