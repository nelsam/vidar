// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package bind contains types that the commander uses to
// identify as types of bindings.  It is kept separate from the
// commander package so that plugins can import it without risking
// rapid changes requiring rebuilds.
//
// For more information, see vidar's plugin package documentation.
package bind

// A Bindable is the base type for all types that can be bound to
// events.
type Bindable interface {
	// Name returns the name of the Bindable.  It must be unique.
	Name() string
}

// A CommandHook is a binding that is bound to a Command.  It is
// up to the Command that the hook is being bound to to define the
// types that may be bound.
//
// The Command that is found by looking up the CommandName() value
// must be a HookedCommand.
type CommandHook interface {
	Bindable

	// CommandName returns the name of the command that this hook
	// wants to bind to.
	CommandName() string
}

// A Command is a binding that happens due to the user explicitly
// requesting an event, usually via a key binding.
type Command interface {
	Bindable

	// Menu returns the name of the menu that the command should be
	// displayed under.
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

// An Executor is a Command that needs to execute some
// operation on one or more elements after it is complete.  It will
// continue to be called for every element currently in the UI until
// it returns a true consume value.
//
// If a true value is never returned for executed, it is assumed that
// the command could not run and is therefor in an errored state.
type Executor interface {
	Command

	Exec(interface{}) Status
}

// A MultiExecutor is a type that makes it easier to execute using
// multiple types in the editor hierarchy.  The MultiExecutor can
// gather types that it needs before Exec is called.
type MultiExecutor interface {
	Command

	// Reset must be implemented to reset the MultiExecutor's state
	Reset()

	// Store is called in the same way that Executor.Exec is called;
	// once for each type in the editor's hierarchy.  Values may be
	// stored by the MultiExecutor to prepare for the final Exec
	// call.
	//
	// When the MultiExecutor has stored enough information to
	// execute, its return value should contain the Executed bit.
	// If every type in the editor's hierarchy is passed to Store
	// without such a return value, it is assumed that the
	// MultiExecutor could not find everything it needs to execute,
	// and Exec will *not* be called.
	Store(interface{}) Status

	// Exec is called after Store has finished.  The return value
	// is a simple error because there are only two states it can
	// be in after running - success or failure.  There is no
	// "waiting" state for a MultiExecutor's Exec method.
	Exec() error
}
