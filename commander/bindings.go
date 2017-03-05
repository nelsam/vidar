// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import "github.com/nelsam/gxui"

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

// A ClonableCommand is a Command that can clone its internal state,
// returning a clone referencing new memory.
type CloneableCommand interface {
	Command

	// Clone creates a clone of the CloneableCommand, returning an
	// exact clone referencing new memory.
	//
	// Clone will be called whenever the state of the CloneableCommand
	// needs to be stored prior to altering its state - especially
	// when the CloneableCommand will be modified by a hook.
	Clone() CloneableCommand
}

// A HookedCommand is a Command that has hooks, which CommandHook
// types may bind to.  It is required to implement CloneableCommand
// because the state of the HookedCommand must be stored before a
// hook causes new hooks to be bound to it.
type HookedCommand interface {
	CloneableCommand

	// Bind takes a CommandHook to bind to the command.  An error
	// should be returned if the CommandHook does not implement
	// any types bindable by the HookedCommand.
	Bind(CommandHook) error
}

// A Starter is a type of Command which needs to initialize itself
// whenever the user wants to run it.
type Starter interface {
	Command

	// Start starts the command.  The element that the command is
	// targeting will be passed in as target.  If the returned
	// status element is non-nil, it will be displayed as an
	// element to display the current status of the command to
	// the user.
	Start(target gxui.Control) (status gxui.Control)
}

// An InputQueue is a type of Command which needs to read user input.
type InputQueue interface {
	Command

	// Next returns the next element for reading user input. By
	// default, this is called every time the commander receives a
	// gxui.KeyEnter event in KeyPress.  If there are situations where
	// this is not the desired behavior, the returned gxui.Focusable
	// can consume the gxui.KeyboardEvent.  If the input element has
	// other keyboard events that would trigger completion, it can
	// implement Completer, which will allow it to define when it
	// is complete.
	//
	// Next will continue to be called until it returns nil, at which
	// point the command is assumed to be done.
	Next() gxui.Focusable
}

// Completer is a type that may optionally be implemented by types
// returned from InputQueue.Next() to decide when they're complete.
// This allows them to consume enter events as newlines or trigger
// completeness off of key presses other than enter.
type Completer interface {
	Command

	// Complete returns whether or not the event signals a completion
	// of the input.
	Complete(gxui.KeyboardEvent) bool
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

// A Statuser is a Bindable that needs to display its status after being
// run.  The bindings should use their discretion for status colors,
// but colors for some common message types are exported by this
// package to keep things consistent.
type Statuser interface {
	Bindable

	// Status returns the element to display for the binding's status.
	// The element will be removed after some time.
	Status() gxui.Control
}

// ColorSetter is a type that can have its color set.
type ColorSetter interface {
	// SetColor is called when a command element is displayed, so
	// that it matches the color theme of the commander.
	SetColor(gxui.Color)
}
