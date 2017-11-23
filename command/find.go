// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/commander/input"
)

type SelectionEditor interface {
	input.Editor
	Controller() *gxui.TextBoxController
	Select(gxui.TextSelectionList)
}

type Find struct {
	driver  gxui.Driver
	theme   *basic.Theme
	editor  SelectionEditor
	display gxui.Label
	pattern *findBox

	next gxui.Focusable
}

func NewFind(driver gxui.Driver, theme *basic.Theme) *Find {
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
	if f.editor == nil {
		return nil
	}
	f.display = f.theme.CreateLabel()
	f.display.SetText("Start typing to search")
	f.pattern = newFindBox(f.driver, f.theme)
	f.next = f.pattern
	f.pattern.OnTextChanged(func([]gxui.TextBoxEdit) {
		f.editor.Controller().ClearSelections()
		needle := f.pattern.Text()
		if len(needle) == 0 {
			f.display.SetText("Start typing to search")
			return
		}
		haystack := f.editor.Text()
		start := 0
		var selections gxui.TextSelectionList

		count := utf8.RuneCountInString(needle)
		length := len(needle)
		pos := 0
		for next := strings.Index(haystack, needle); next != -1; next = strings.Index(haystack[start:], needle) {
			pos += utf8.RuneCountInString(haystack[start : start+next])
			selection := gxui.CreateTextSelection(pos, pos+count, false)
			selections = append(selections, selection)
			pos += count
			start += (next + length)
		}
		f.editor.Select(selections)
		f.display.SetText(fmt.Sprintf("%s: %d results found", needle, len(selections)))
	})
	return f.display
}

func (f *Find) Name() string {
	return "find"
}

func (f *Find) Menu() string {
	return "Edit"
}

func (f *Find) Next() gxui.Focusable {
	next := f.next
	f.next = nil
	return next
}

type findBox struct {
	mixins.TextBox
}

func newFindBox(driver gxui.Driver, theme *basic.Theme) *findBox {
	box := &findBox{}
	box.Init(driver, theme)
	return box
}

func (b *findBox) Init(driver gxui.Driver, theme *basic.Theme) {
	b.TextBox.Init(b, driver, theme, theme.DefaultMonospaceFont())

	b.SetTextColor(theme.TextBoxDefaultStyle.FontColor)
	b.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	b.SetPadding(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	b.SetBackgroundBrush(theme.TextBoxDefaultStyle.Brush)
	b.SetDesiredWidth(math.MaxSize.W)
	b.SetMultiline(false)
}

func findEditor(elem interface{}) SelectionEditor {
	switch src := elem.(type) {
	case SelectionEditor:
		return src
	case commander.Elementer:
		for _, child := range src.Elements() {
			if editor := findEditor(child); editor != nil {
				return editor
			}
		}
	}
	return nil
}
