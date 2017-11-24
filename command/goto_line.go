// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"unicode"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/status"
)

type Scroller interface {
	LineStart(int) int
	ScrollToLine(int)
}

type LineControl interface {
	LineCount() int
	SetCaret(int)
}

type GotoLine struct {
	status.General

	lineNumInput gxui.TextBox
	input        gxui.Focusable

	editor Scroller
	ctrl   LineControl
}

func NewGotoLine(theme gxui.Theme) *GotoLine {
	input := theme.CreateTextBox()
	input.OnTextChanged(func([]gxui.TextBoxEdit) {
		runes := []rune(input.Text())
		for index := 0; index < len(runes); index++ {
			if !unicode.IsDigit(runes[index]) {
				runes = append(runes[:index], runes[index+1:]...)
				index--
			}
		}
		text := string(runes)
		if text != input.Text() {
			input.SetText(text)
		}
	})
	g := &GotoLine{}
	g.Theme = theme
	g.lineNumInput = input
	return g
}

func (g *GotoLine) Start(on gxui.Control) gxui.Control {
	g.lineNumInput.SetText("")
	g.input = g.lineNumInput
	return nil
}

func (g *GotoLine) Name() string {
	return "goto-line"
}

func (g *GotoLine) Menu() string {
	return "Edit"
}

func (g *GotoLine) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyG,
	}}
}

func (g *GotoLine) Next() gxui.Focusable {
	input := g.input
	g.input = nil
	return input
}

func (g *GotoLine) Reset() {
	g.editor = nil
	g.ctrl = nil
}

func (g *GotoLine) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case LineControl:
		g.ctrl = src
	case Scroller:
		g.editor = src
	}
	if g.editor != nil && g.ctrl != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (g *GotoLine) Exec() error {
	lineStr := g.lineNumInput.Text()
	if lineStr == "" {
		g.Warn = "No line number provided"
		return nil
	}
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		// This shouldn't ever happen, but in the interests of avoiding data loss,
		// we just log that it did.
		log.Printf("ERR: goto-line: failed to parse %s as a line number", g.lineNumInput.Text())
		return err
	}
	line-- // Convert to zero-based.

	if line >= g.ctrl.LineCount() {
		// TODO: Should we just choose the last line?
		g.Err = fmt.Sprintf("Line %d is past the end of the file", line)
		return fmt.Errorf("%d is too large", line)
	}
	if line == -1 {
		// TODO: Should we just choose the first line?
		g.Err = "Line 0 does not exist"
		return errors.New("Invalid line")
	}
	g.ctrl.SetCaret(g.editor.LineStart(line))
	g.editor.ScrollToLine(line)
	return nil
}
