// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui_playground/controller"
)

type commandBox struct {
	mixins.LinearLayout

	driver     gxui.Driver
	theme      *basic.Theme
	controller Controller

	label   gxui.Label
	current controller.Command
	display gxui.Control
	input   gxui.Focusable
}

func NewCommandBox(theme *basic.Theme, controller Controller) CommandBox {
	box := &commandBox{}
	box.Init(theme, controller)
	return box
}

func (b *commandBox) Init(theme *basic.Theme, controller Controller) {
	b.theme = theme
	b.controller = controller
	b.label = b.theme.CreateLabel()

	b.LinearLayout.Init(b, b.theme)
	b.SetDirection(gxui.LeftToRight)
	size := b.theme.DefaultMonospaceFont().GlyphMaxSize()
	size.W = math.MaxSize.W
	b.SetSize(size)
}

func (b *commandBox) Clear() {
	b.label.SetText("none")
	b.clearDisplay()
	b.clearInput()
	b.current = nil
}

func (b *commandBox) Run(command controller.Command) (needsInput bool) {
	b.current = command

	b.label.SetText(b.current.Name())
	b.display = b.current.Start(b.controller)
	if b.display != nil {
		b.AddChild(b.display)
	}
	return b.nextInput()
}

func (b *commandBox) Current() controller.Command {
	return b.current
}

func (b *commandBox) KeyPress(event gxui.KeyboardEvent) bool {
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
