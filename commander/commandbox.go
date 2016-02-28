// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
)

var (
	cmdColor = gxui.Color{
		R: 0.3,
		G: 0.3,
		B: 0.6,
		A: 1,
	}
	displayColor = gxui.Color{
		R: 0.3,
		G: 1,
		B: 0.6,
		A: 1,
	}
)

// Completer is a type which defines when a gxui.KeyboardEvent
// completes an action.  Types returned from
// "../commands".Command.Next() may implement this if they don't want
// to immediately complete when the "enter" key is pressed.
type Completer interface {
	// Complete returns whether or not the key signals a completion of
	// the input.
	Complete(gxui.KeyboardEvent) bool
}

// A Command is any command that can be run by the commandBox.
type Command interface {
	// Start starts the command.  The returned gxui.Control can be
	// used to display the status of the command.
	//
	// If no display element is necessary, or if display will be taken
	// care of by the input elements, Start should return nil.
	Start(gxui.Control) gxui.Control

	// Name returns the name of the command
	Name() string

	// Next returns the next control element for reading user input.
	// By default, this is called every time the commander receives a
	// gxui.KeyEnter event in KeyPress.  If there are situations where
	// this is not the desired behavior, the returned gxui.Focusable
	// can consume the event.  If more keys are required to signal the
	// end of input, the gxui.Focusable can implement Completer.
	//
	// Next will continue to be called until it returns nil, at which
	// point the command is assumed to be done.
	Next() gxui.Focusable

	// Exec implements controller.Executor
	Exec(interface{}) (executed, consume bool)
}

type colorSetter interface {
	SetColor(gxui.Color)
}

type commandBox struct {
	mixins.LinearLayout

	controller Controller

	label   gxui.Label
	current Command
	display gxui.Control
	input   gxui.Focusable
}

func newCommandBox(theme gxui.Theme, controller Controller) *commandBox {
	box := &commandBox{}

	box.controller = controller
	box.label = theme.CreateLabel()
	box.label.SetColor(cmdColor)

	box.LinearLayout.Init(box, theme)
	box.SetDirection(gxui.LeftToRight)
	box.AddChild(box.label)
	box.Clear()
	return box
}

func (b *commandBox) Clear() {
	b.label.SetText("none")
	b.clearDisplay()
	b.clearInput()
	b.current = nil
}

func (b *commandBox) Run(command Command) (needsInput bool) {
	b.current = command

	b.label.SetText(b.current.Name())
	b.display = b.current.Start(b.controller)
	if b.display != nil {
		if colorSetter, ok := b.display.(colorSetter); ok {
			colorSetter.SetColor(displayColor)
		}
		b.AddChild(b.display)
	}
	return b.nextInput()
}

func (b *commandBox) Current() Command {
	return b.current
}

func (b *commandBox) KeyPress(event gxui.KeyboardEvent) (consume bool) {
	if event.Modifier == 0 && event.Key == gxui.KeyEscape {
		return false
	}
	isEnter := event.Modifier == 0 && event.Key == gxui.KeyEnter
	complete := isEnter
	if completer, ok := b.input.(Completer); ok {
		complete = completer.Complete(event)
	}
	if complete {
		hasMore := b.nextInput()
		complete = !hasMore
	}
	return !(complete && isEnter)
}

func (b *commandBox) HasFocus() bool {
	if b.input == nil {
		return false
	}
	return b.input.HasFocus()
}

func (b *commandBox) clearDisplay() {
	if b.display != nil {
		b.RemoveChild(b.display)
		b.display = nil
	}
}

func (b *commandBox) clearInput() {
	if b.input != nil {
		b.RemoveChild(b.input)
		b.input = nil
	}
}

type debuggable interface {
	SetDebug(bool)
}

func (b *commandBox) nextInput() (more bool) {
	next := b.current.Next()
	if next == nil {
		return false
	}
	b.clearInput()
	b.input = next
	b.AddChild(b.input)
	gxui.SetFocus(b.input)
	return true
}
