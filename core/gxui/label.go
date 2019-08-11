// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/theme"
	"github.com/nelsam/vidar/ui"
)

type Label struct {
	label gxui.Label
}

func (l Label) Control() gxui.Control {
	return l.label
}

func (l Label) SetText(r []rune) {
	l.label.SetText(string(r))
}

func (l Label) Text() []rune {
	return []rune(l.label.Text())
}

func (l Label) SetSpans(spans ...ui.Span) error {
	if len(spans) == 0 {
		return nil
	}
	span := spans[0]
	color := span.Highlight.Foreground
	l.label.SetColor(gxui.Color(color))
	if len(spans) > 1 || span.Start > 0 || int(span.Length) < len(l.label.Text()) {
		return fmt.Errorf("Label types cannot apply more than one color")
	}
	return nil
}

func (l Label) Spans() []ui.Span {
	return []ui.Span{
		{Highlight: theme.Highlight{Foreground: theme.Color(l.label.Color())}, Start: 0, Length: uint64(len(l.Text()))},
	}
}
