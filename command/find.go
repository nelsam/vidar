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
	"github.com/nelsam/vidar/commander/text"
)

type SelectionEditor interface {
	text.Editor
	Controller() *gxui.TextBoxController
	SelectSlice([]gxui.TextSelection)
	ScrollToRune(int)
}

type Find struct {
	mixins.LinearLayout

	driver     gxui.Driver
	theme      *basic.Theme
	editor     SelectionEditor
	display    gxui.Label
	pattern    *findBox
	prevS      gxui.Button
	nextS      gxui.Button
	selections []int
	selection  int
}

func NewFind(driver gxui.Driver, theme *basic.Theme) *Find {
	finder := &Find{}
	finder.Init(driver, theme)
	return finder
}

func (f *Find) Init(driver gxui.Driver, theme *basic.Theme) {
	f.LinearLayout.Init(f, theme)
	f.SetDirection(gxui.RightToLeft)
	f.driver = driver
	f.theme = theme

	f.display = f.theme.CreateLabel()
	f.display.SetText("Start typing to search")

	f.prevS = f.theme.CreateButton()
	f.prevS.SetText("<")
	f.prevS.OnClick(func(ev gxui.MouseEvent) {
		if len(f.selections) != 0 {
			f.selection = getNext(f.selection, len(f.selections), -1)
			f.editor.ScrollToRune(f.selections[f.selection])
		}
	})

	f.nextS = f.theme.CreateButton()
	f.nextS.SetText(">")
	f.nextS.OnClick(func(ev gxui.MouseEvent) {
		if len(f.selections) != 0 {
			f.selection = getNext(f.selection, len(f.selections), 1)
			f.editor.ScrollToRune(f.selections[f.selection])
		}
	})
	f.AddChild(f.nextS)
	f.AddChild(f.prevS)
}

func (f *Find) KeyPress(event gxui.KeyboardEvent) bool {
	if event.Modifier == gxui.ModControl {
		if event.Key == gxui.KeyN {
			f.nextS.Click(gxui.MouseEvent{})
		} else if event.Key == gxui.KeyP {
			f.prevS.Click(gxui.MouseEvent{})
		}
	}
	return f.pattern.KeyPress(event)
}

func (f *Find) KeyDown(event gxui.KeyboardEvent) {
	f.pattern.KeyDown(event)
}

func (f *Find) KeyUp(event gxui.KeyboardEvent) {
	f.pattern.KeyUp(event)
}

func (f *Find) KeyStroke(event gxui.KeyStrokeEvent) bool {
	return f.pattern.KeyStroke(event)
}

func (f *Find) KeyRepeat(event gxui.KeyboardEvent) {
	f.pattern.KeyRepeat(event)
}

func (f *Find) Paint(c gxui.Canvas) {
	f.LinearLayout.Paint(c)

	if f.HasFocus() {
		r := f.Size().Rect()
		s := f.theme.FocusedStyle
		c.DrawRoundedRect(r, 3, 3, 3, 3, s.Pen, s.Brush)
	}
}

func (f *Find) IsFocusable() bool {
	return f.pattern.IsFocusable()
}

func (f *Find) HasFocus() bool {
	return f.pattern.HasFocus()
}

func (f *Find) GainedFocus() {
	f.pattern.GainedFocus()
}

func (f *Find) LostFocus() {
	f.pattern.LostFocus()
}

func (f *Find) OnGainedFocus(callback func()) gxui.EventSubscription {
	return f.pattern.OnGainedFocus(callback)
}

func (f *Find) OnLostFocus(callback func()) gxui.EventSubscription {
	return f.pattern.OnLostFocus(callback)
}

func (f *Find) Start(control gxui.Control) gxui.Control {
	f.editor = findEditor(control)
	if f.editor == nil {
		return nil
	}
	f.pattern = newFindBox(f.driver, f.theme)
	f.AddChild(f.pattern)

	f.pattern.OnTextChanged(func([]gxui.TextBoxEdit) {
		f.editor.Controller().ClearSelections()
		needle := f.pattern.Text()
		if len(needle) == 0 {
			f.display.SetText("Start typing to search")
			return
		}
		f.selections = []int{}

		haystack := f.editor.Text()
		start := 0
		var selections []gxui.TextSelection

		count := utf8.RuneCountInString(needle)
		length := len(needle)
		pos := 0
		for next := strings.Index(haystack, needle); next != -1; next = strings.Index(haystack[start:], needle) {
			pos += utf8.RuneCountInString(haystack[start : start+next])
			selection := gxui.CreateTextSelection(pos, pos+count, false)
			selections = append(selections, selection)
			f.selections = append(f.selections, pos)
			pos += count
			start += (next + length)
		}
		f.selection = len(f.selections) - 1
		f.editor.SelectSlice(selections)
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

func (f *Find) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyF,
	}}
}

func (f *Find) Next() gxui.Focusable {
	return f
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

func getNext(i, l, s int) int {
	if s > 0 {
		return (i + s) % l
	}
	return (i + l + s) % l
}
