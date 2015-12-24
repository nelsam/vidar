// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commands"
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

type colorSetter interface {
	SetColor(gxui.Color)
}

type commandBox struct {
	mixins.LinearLayout

	driver     gxui.Driver
	theme      *basic.Theme
	controller Controller

	label   gxui.Label
	current commands.Command
	// TODO: think about moving this to current.Done()
	currentDone bool
	display     gxui.Control
	input       gxui.Focusable
}

func newCommandBox(theme *basic.Theme, controller Controller) *commandBox {
	box := &commandBox{}
	box.Init(theme, controller)
	return box
}

func (b *commandBox) Init(theme *basic.Theme, controller Controller) {
	b.theme = theme
	b.controller = controller
	b.label = b.theme.CreateLabel()
	b.label.SetColor(cmdColor)

	b.LinearLayout.Init(b, b.theme)
	b.SetDirection(gxui.LeftToRight)
	b.AddChild(b.label)
	b.Clear()
}

func (b *commandBox) Clear() {
	b.label.SetText("none")
	b.clearDisplay()
	b.clearInput()
	b.current = nil
}

func (b *commandBox) Run(command commands.Command) (needsInput bool) {
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

// Done is a workaround for the fact that events don't get consumed in the way
// I was expecting them to.  Currently, the Commander still ends up getting
// the events, even when they're consumed.  I don't know why.
func (b *commandBox) Done() bool {
	return b.currentDone
}

func (b *commandBox) Current() commands.Command {
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
	b.currentDone = complete
	// TODO: figure out why this isn't consuming the event when it should.  The commander
	// still receives this event, even when it's consumed.
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
		b.currentDone = true
		return false
	}
	b.currentDone = false
	b.clearInput()
	b.input = next
	b.AddChild(b.input)
	gxui.SetFocus(b.input)
	return true
}
