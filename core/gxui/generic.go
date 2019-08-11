// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import "github.com/nelsam/gxui"

// Control is a generic ui.Wrapper for gxui control types.  It is
// here for transitionary periods while we decouple core from gxui.
type Control struct {
	Elem gxui.Control
}

func (c Control) Control() gxui.Control {
	return c.Elem
}
