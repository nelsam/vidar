// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"time"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
)

const maxStatusAge = 5 * time.Second

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

	ColorErr = gxui.Color{
		R: 1.0,
		G: 0.2,
		B: 0,
		A: 1,
	}
	ColorWarn = gxui.Color{
		R: 0.8,
		G: 0.7,
		B: 0.1,
		A: 1,
	}
	ColorInfo = gxui.Color{
		R: 0.1,
		G: 1,
		B: 0,
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

// A Command is any command that can be run by the commandBox.  Commands
// have many optional interfaces they can implement - see other
// interfaces in this package for what else can be implemented by a
// command.
type Command interface {
	// Name returns the name of the command
	Name() string
}

// A Starter is a type of command which needs to initialize itself
// whenever the command is started.
type Starter interface {
	// Start starts the command.  The element that the command is
	// targetting will be passed in as target.  If the returned
	// status element is non-nil, it will be displayed as an
	// element to display the current status of the command to
	// the user.
	Start(target gxui.Control) (status gxui.Control)
}

// An InputQueue is a type of command which needs to read user input
// before executing a command.
type InputQueue interface {
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

// An Executor is a type of command that needs to execute some
// operation on one or more elements.  It will continue to be
// called for every element currently in the UI until it returns
// true for consume.
//
// If executed is never returned as true, an error will be
// displayed to the user stating that the command never
// executed.
type Executor interface {
	Exec(interface{}) (executed, consume bool)
}

// A Statuser is a type that needs to display its status after being
// run.  The commands should use their discretion for status colors,
// but colors for some common message types are exported by this
// package to keep things consistent.
type Statuser interface {
	// Status returns the element to display for the command's status.
	// The element will be removed after some time.
	Status() gxui.Control
}

// ColorSetter is a type that can have its color set.
type ColorSetter interface {
	// SetColor is called when a command element is displayed, so
	// that it matches the color theme of the commander.
	SetColor(gxui.Color)
}

type commandBox struct {
	mixins.LinearLayout

	driver     gxui.Driver
	controller Controller

	label   gxui.Label
	current Command
	display gxui.Control
	input   gxui.Focusable
	status  gxui.Control

	statusTimer *time.Timer
}

func newCommandBox(driver gxui.Driver, theme gxui.Theme, controller Controller) *commandBox {
	box := &commandBox{
		driver:     driver,
		controller: controller,
	}

	box.label = theme.CreateLabel()
	box.label.SetColor(cmdColor)

	box.LinearLayout.Init(box, theme)
	box.SetDirection(gxui.LeftToRight)
	box.AddChild(box.label)
	box.Clear()
	return box
}

func (b *commandBox) Finish() {
	statuser, ok := b.current.(Statuser)
	if !ok {
		b.Clear()
		return
	}
	b.status = statuser.Status()
	if b.status == nil {
		b.Clear()
		return
	}
	b.clearDisplay()
	b.clearInput()
	b.AddChild(b.status)
	b.statusTimer = time.AfterFunc(maxStatusAge, func() {
		b.driver.CallSync(func() {
			b.Clear()
		})
	})
}

func (b *commandBox) Clear() {
	b.label.SetText("none")
	b.clearDisplay()
	b.clearInput()
	b.clearStatus()
	b.current = nil
}

func (b *commandBox) Run(command Command) (needsInput bool) {
	if b.statusTimer != nil {
		b.statusTimer.Stop()
	}
	b.current = command

	b.label.SetText(b.current.Name())
	b.startCurrent()
	return b.nextInput()
}

func (b *commandBox) startCurrent() {
	starter, ok := b.current.(Starter)
	if !ok {
		return
	}
	b.display = starter.Start(b.controller)
	if b.display == nil {
		return
	}
	if colorSetter, ok := b.display.(ColorSetter); ok {
		colorSetter.SetColor(displayColor)
	}
	b.AddChild(b.display)
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
	if b.display == nil {
		return
	}
	b.RemoveChild(b.display)
	b.display = nil
}

func (b *commandBox) clearInput() {
	if b.input == nil {
		return
	}
	b.RemoveChild(b.input)
	b.input = nil
}

func (b *commandBox) clearStatus() {
	if b.status == nil {
		return
	}
	b.RemoveChild(b.status)
	b.status = nil
}

func (b *commandBox) nextInput() (more bool) {
	queue, ok := b.current.(InputQueue)
	if !ok {
		return false
	}
	next := queue.Next()
	if next == nil {
		return false
	}
	b.clearInput()
	b.input = next
	b.AddChild(b.input)
	gxui.SetFocus(b.input)
	return true
}
