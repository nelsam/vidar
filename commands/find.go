// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"fmt"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
)

type EditorFinder interface {
	CurrentEditor() gxui.CodeEditor
}

type Find struct {
	driver  gxui.Driver
	theme   *basic.Theme
	editor  gxui.CodeEditor
	display gxui.Label
	pattern *findBox

	next gxui.Focusable
}

func NewFinder(driver gxui.Driver, theme *basic.Theme) *Find {
	finder := &Find{}
	finder.Init(driver, theme)
	return finder
}

func (f *Find) Init(driver gxui.Driver, theme *basic.Theme) {
	f.driver = driver
	f.theme = theme
}

func (f *Find) Start(control gxui.Control) gxui.Control {
	f.editor = findEditor(control)
	f.display = f.theme.CreateLabel()
	f.display.SetText("Start typing to search")
	f.pattern = newFindBox(f.driver, f.theme, f.editor)
	f.next = f.pattern
	f.pattern.OnTextChanged(func([]gxui.TextBoxEdit) {
		f.editor.Controller().ClearSelections()
		needle := f.pattern.Text()
		if len(needle) == 0 {
			f.display.SetText("Start typing to search")
			return
		}
		haystack := f.editor.Text()
		moveCursor := true
		start := 0
		for next := strings.Index(haystack, needle); next != -1; next = strings.Index(haystack[start:], needle) {
			start += next
			selection := gxui.CreateTextSelection(start, start+len(needle), moveCursor)
			moveCursor = false
			f.editor.Controller().AddSelection(selection)
			start++
		}
		f.display.SetText(fmt.Sprintf("%s: %d results found", needle, f.editor.Controller().SelectionCount()))
	})
	return f.display
}

func (f *Find) Name() string {
	return "find"
}

func (f *Find) Next() gxui.Focusable {
	next := f.next
	f.next = nil
	return next
}

// Exec is a noop on a Find command.  It is unnecessary because Find will be
// searching as a user types.
func (f *Find) Exec(target interface{}) (executed, consume bool) {
	return true, true
}

type findBox struct {
	mixins.TextBox
	editor gxui.CodeEditor
}

func newFindBox(driver gxui.Driver, theme *basic.Theme, editor gxui.CodeEditor) *findBox {
	box := &findBox{}
	box.Init(driver, theme, editor)
	return box
}

func (b *findBox) Init(driver gxui.Driver, theme *basic.Theme, editor gxui.CodeEditor) {
	b.TextBox.Init(b, driver, theme, theme.DefaultMonospaceFont())
	b.editor = editor

	b.SetTextColor(theme.TextBoxDefaultStyle.FontColor)
	b.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	b.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	b.SetBackgroundBrush(theme.TextBoxDefaultStyle.Brush)
	b.SetDesiredWidth(math.MaxSize.W)
	b.SetMultiline(false)
}

func findEditor(control gxui.Control) gxui.CodeEditor {
	switch src := control.(type) {
	case EditorFinder:
		return src.CurrentEditor()
	case gxui.Parent:
		for _, child := range src.Children() {
			if editor := findEditor(child.Control); editor != nil {
				return editor
			}
		}
	}
	return nil
}
