// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/editor"
)

// A ClonableCommand is a Command that can clone its internal state,
// returning a clone referencing new memory.
type CloneableCommand interface {
	bind.Command

	// Clone creates a clone of the CloneableCommand, returning an
	// exact clone referencing new memory.
	//
	// Clone will be called whenever the state of the CloneableCommand
	// needs to be stored prior to altering its state - especially
	// when the CloneableCommand will be modified by a hook.
	Clone() CloneableCommand
}

// InputHandler handles key strokes from the keyboard.  An InputHandler
// is essentially the text engine for the text editor.  It takes each
// character entered by the user and processes it into text (or commands)
// to provide to the editor.  If you want vidar to feel like vim, this
// is the type of command to implement.
//
// It is the responsibility of the input handler to correctly deal with
// text input fast enough to keep up with a user.  Because of this,
// it's recommended that any alternate input engines are careful about
// how slow hooks are allowed to be.
//
// Since the InputHandler will handle all edits to the text, it should
// accept both commands.ChangeHook and commands.ContextChangeHook as
// hooks that can bind to it.
type InputHandler interface {
	bind.Bindable
	Clone() InputHandler
	Bind(bind.CommandHook) error
	HandleInput(focused *editor.CodeEditor, stroke gxui.KeyStrokeEvent)
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
	Bind(bind.CommandHook) error
}

// A Starter is a type of Command which needs to initialize itself
// whenever the user wants to run it.
type Starter interface {
	bind.Command

	// Start starts the command.  The element that the command is
	// targeting will be passed in as target.  If the returned
	// status element is non-nil, it will be displayed as an
	// element to display the current status of the command to
	// the user.
	Start(target gxui.Control) (status gxui.Control)
}

// An InputQueue is a type of Command which needs to read user input.
type InputQueue interface {
	bind.Command

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
	bind.Command

	// Complete returns whether or not the event signals a completion
	// of the input.
	Complete(gxui.KeyboardEvent) bool
}

// A Statuser is a Bindable that needs to display its status after being
// run.  The bindings should use their discretion for status colors,
// but colors for some common message types are exported by this
// package to keep things consistent.
type Statuser interface {
	bind.Bindable

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
