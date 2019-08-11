// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
)

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
