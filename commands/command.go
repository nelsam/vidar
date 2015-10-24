// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package commands defines and implements commands which can be run
// against Vidar's controller.  Commands are designed to be bound to
// events - key presses, mouse clicks, or commands in the command box.
package commands

import "github.com/nelsam/gxui"

// Command is a command that executes against the Controller or one of
// its children.
type Command interface {
	// Start starts the command.  The returned gxui.Control can be
	// used to display the status of the command.  The elements
	// returned by Next() should have OnUnfocus() triggers to update
	// this element.
	//
	// If no display element is necessary, or if display will be taken
	// care of by the input elements, Start should return nil.
	Start(gxui.Control) gxui.Control

	// Name returns the name of the command
	Name() string

	// Next returns the next control element for reading user input.
	// By default, this is called every time the commander receives a
	// gxui.KeyEnter event in KeyPress.  If there are situations where
	// this is not the desired behavior, the gxui.Control should
	// consume the event.  If more keys are required to signal the end
	// of input, the control should implement Completer.
	//
	// When no more user input is required, Next should return nil.
	Next() gxui.Focusable

	// Exec executes the command on the provided element.  The return
	// values represent whether or not the command was executed, and
	// whether or not the command needs to continue descending through
	// the elements.
	//
	// If a command is run against all elements in Vidar but never
	// returns executed == true, it will be considered an error and
	// the error will be presented to the user.
	//
	// As soon as consume returns true, the command will be considered
	// complete and will not be run against any more elements.
	Exec(interface{}) (executed, consume bool)
}
