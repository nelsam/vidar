// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package lsp

import (
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
)

// Binder is a type that keeps track of all lsp connections and
// returns Bindables when files are opened.  Any language-specific
// language servers should be registered as hooks to Binder.
type Binder struct {
}

func (*Binder) Name() string {
	return "lsp-client"
}

func (*Binder) OpName() string {
	return "focus-location"
}

// EditorChanged implements focus.EditorChanger
func (c *Conn) EditorChanged(e input.Editor) {
	c.curr = e
}

func (b *Binder) Bind(h bind.Bindable) (bind.HookedOp, error) {
}
