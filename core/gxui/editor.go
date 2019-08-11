// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/vidar/theme"
	"github.com/nelsam/vidar/ui"
)

type Editor struct {
	ed gxui.CodeEditor
}

func (e Editor) Control() gxui.Control {
	return e.ed
}

func (e Editor) SetText(r []rune) {
	e.ed.Controller().SetTextRunes(r)
}

func (e Editor) Text() []rune {
	return e.ed.Controller().TextRunes()
}

func (e Editor) SetCarets(pos ...int) error {
	l := len(e.ed.Controller().TextRunes())
	var sel []gxui.TextSelection
	for _, c := range pos {
		if c < 0 {
			return fmt.Errorf("gxui: cannot apply carets: %d is before the start of the file (0)", c)
		}
		if c > l {
			return fmt.Errorf("gxui: cannot apply carets: %d is past the end of the file (%d)", c, l)
		}
		sel = append(sel, gxui.CreateTextSelection(c, c, false))
	}
	e.ed.Controller().SetSelections(sel)
	return nil
}

func (e Editor) Carets() []int {
	return e.ed.Carets()
}

func (e Editor) SetSpans(l ...ui.Span) error {
	var layers gxui.CodeSyntaxLayers
	m := make(map[theme.Highlight]*gxui.CodeSyntaxLayer)
	for _, s := range l {
		layer := m[s.Highlight]
		if layer == nil {
			layer = gxui.CreateCodeSyntaxLayer()
			layer.SetColor(gxui.Color(s.Highlight.Foreground))
			layer.SetBackgroundColor(gxui.Color(s.Highlight.Background))
			layers = append(layers, layer)
			m[s.Highlight] = layer
		}
		layer.Add(int(s.Start), int(s.Length))

	}
	e.ed.SetSyntaxLayers(layers)
	return nil
}

func (e Editor) Spans() []ui.Span {
	var s []ui.Span
	for _, l := range e.ed.SyntaxLayers() {
		fg := l.Color()
		if fg == nil {
			fg = &gxui.Color{}
		}
		bg := l.BackgroundColor()
		if bg == nil {
			bg = &gxui.Color{}
		}
		for _, span := range l.Spans() {
			start, end := span.Span()
			uispan := ui.Span{
				Highlight: theme.Highlight{
					Foreground: theme.Color(*fg),
					Background: theme.Color(*bg),
				},
				Start:  uint64(start),
				Length: uint64(end - start),
			}
			s = append(s, uispan)
		}
	}
	return s
}

type codeEditor struct {
	mixins.CodeEditor
}

func (e *codeEditor) DataChanged(recreate bool) {
	e.List.DataChanged(recreate)
}

func (e *codeEditor) CreateLine(theme gxui.Theme, index int) (mixins.TextBoxLine, gxui.Control) {
	lineNumber := theme.CreateLabel()
	// TODO: set the color a bit more gray
	lineNumber.SetText(fmt.Sprintf("%4d", index+1))
	lineNumber.SetMargin(math.Spacing{L: 0, T: 0, R: 3, B: 0})

	line := &mixins.CodeEditorLine{}
	line.Init(line, theme, &e.CodeEditor, index)

	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	layout.AddChild(lineNumber)
	layout.AddChild(line)

	return line, layout
}
