// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/plugin/ui"
)

type Creator struct {
	theme gxui.Theme
}

func (c *Creator) Parent() interface{} {
	return c.theme
}

func (c *Creator) Label(text string) ui.Element {
	l := c.theme.CreateLabel()
	l.SetText(text)
	return &LabelElement{l}
}

type LabelElement struct {
	label gxui.Label
}

func (l *LabelElement) Foreground() ui.Color {
	return ui.Color(l.label.Color())
}

func (l *LabelElement) SetForeground(c ui.Color) {
	l.label.SetColor(gxui.Color(c))
}
