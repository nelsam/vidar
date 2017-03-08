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

// An Executor is a Command that needs to execute some
// operation on one or more elements after it is complete.  It will
// continue to be called for every element currently in the UI until
// it returns a true consume value.
//
// If a true value is never returned for executed, it is assumed that
// the command could not run and is therefor in an errored state.
type Executor interface {
	Command

	Exec(interface{}) (executed, consume bool)
}
