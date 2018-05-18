// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import "github.com/nelsam/gxui"

// window is a slightly modified gxui.Window that we use to wrap up
// some child types.
type window struct {
	gxui.Window
	child interface{}
}

func newWindow(t gxui.Theme) *window {
	return &window{
		Window: t.CreateWindow(1600, 800, "Vidar - GXUI Go Editor"),
	}
}

func (w *window) Elements() []interface{} {
	return []interface{}{w.child}
}
