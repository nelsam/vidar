// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
)

// Edit is a type containing details about edited text.
type Edit struct {
	At  int
	Old []rune
	New []rune
}

// Handler handles key strokes from the keyboard.  A Handler
// is essentially the text engine for the text editor.  It takes each
// character entered by the user and processes it into text (or commands)
// to provide to the editor.  If you want vidar to feel like vim, this
// is the type of command to implement.
//
// It is the responsibility of the input handler to correctly deal with
// text input fast enough to keep up with a user.  When implementing a
// Handler, it's important to use goroutines to run hooks and make sure
// that slow hooks are not slowing down the Handler or other hooks.
//
// Since the Handler will replace vidar's default Handler, it should
// accept both commands.ChangeHook and commands.ContextChangeHook as
// hooks that can bind to it in order to support hooks that intend to
// bind to vidar's default Handler.
type Handler interface {
	bind.Bindable
	New() Handler
	Init(Editor, []rune)
	Bind(bind.CommandHook) error
	Apply(focused Editor, edits ...Edit)
	HandleEvent(focused Editor, ev gxui.KeyboardEvent)
	HandleInput(focused Editor, stroke gxui.KeyStrokeEvent)
}
