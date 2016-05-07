// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"log"
	"strconv"
	"unicode"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
)

type GotoLine struct {
	editor       *editor.CodeEditor
	lineNumInput gxui.TextBox
	input        gxui.Focusable
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
	return &GotoLine{
		lineNumInput: input,
	}
}

func (g *GotoLine) Start(on gxui.Control) gxui.Control {
	g.editor = findEditor(on)
	if g.editor == nil {
		return nil
	}
	g.lineNumInput.SetText("")
	g.input = g.lineNumInput
	return nil
}

func (g *GotoLine) Name() string {
	return "goto-line"
}

func (g *GotoLine) Next() gxui.Focusable {
	input := g.input
	g.input = nil
	return input
}

func (g *GotoLine) Exec(on interface{}) (executed, consume bool) {
	lineStr := g.lineNumInput.Text()
	if lineStr == "" {
		return true, true
	}
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		log.Printf("goto-line: failed to parse %s as a line number", g.lineNumInput.Text())
		return true, true
	}
	line = oneToZeroBased(line)
	if line >= g.editor.Controller().LineCount() {
		return true, true
	}
	g.editor.Controller().SetCaret(g.editor.LineStart(line))
	g.editor.ScrollToLine(line)
	return true, true
}
