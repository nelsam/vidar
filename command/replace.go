// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
)

type Replace struct {
	driver gxui.Driver
	theme  *basic.Theme

	find      *findBox
	replace   *findBox
	status    gxui.Label
	editor    SelectionEditor
	applier   Applier
	edits     []input.Edit
	is_select bool

	input <-chan gxui.Focusable
}

func NewReplace(driver gxui.Driver, theme *basic.Theme) *Replace {
	replacer := &Replace{}
	replacer.Init(driver, theme)
	return replacer
}

func (f *Replace) Init(driver gxui.Driver, theme *basic.Theme) {
	f.driver = driver
	f.theme = theme

}

func (f *Replace) Start(control gxui.Control) gxui.Control {
	f.editor = findEditor(control)
	if f.editor == nil {
		return nil
	}
	f.is_select = false

	f.find = newFindBox(f.driver, f.theme)
	f.find.OnTextChanged(func([]gxui.TextBoxEdit) {
		f.editor.Controller().ClearSelections()

		f.is_select = true
		needle := f.find.Text()
		if len(needle) == 0 {
			f.status.SetText("Start typing to search")
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
		f.status.SetText(fmt.Sprintf("%s: %d results found", needle, len(selections)))
	})
	f.find.OnKeyPress(func(ev gxui.KeyboardEvent) {
		switch ev.Key {
		case gxui.KeyEscape:
			if f.is_select {
				f.editor.Controller().ClearSelections()
			}
		}
	})

	f.replace = newFindBox(f.driver, f.theme)
	f.replace.OnKeyPress(func(ev gxui.KeyboardEvent) {
		switch ev.Key {
		case gxui.KeyEnter:
			f.edits = []input.Edit{}
			selections := f.editor.Controller().Selections()

			for i := selections.Len(); i != 0; i-- {
				begin, end := selections.Interval(i - 1)
				str := []rune(f.editor.Text())[begin:end]
				f.edits = append(f.edits, input.Edit{
					At:  int(begin),
					Old: str,
					New: []rune(f.replace.Text()),
				})
			}
			f.editor.Controller().ClearSelections()
		case gxui.KeyEscape:
			f.editor.Controller().ClearSelections()
		}
	})
	f.status = f.theme.CreateLabel()

	input := make(chan gxui.Focusable, 3)
	f.input = input
	input <- f.find
	input <- f.replace
	close(input)

	return f.status
}

func (f *Replace) Name() string {
	return "replace"
}

func (f *Replace) Menu() string {
	return "Edit"
}

func (f *Replace) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyR,
	}}
}

func (c *Replace) Reset() {
	c.editor = nil
	c.applier = nil
}

func (c *Replace) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case SelectionEditor:
		c.editor = src
	case Applier:
		c.applier = src
	}
	if c.editor != nil && c.applier != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (f *Replace) Exec() error {
	if len(f.edits) != 0 {
		f.applier.Apply(f.editor, f.edits...)
	}

	return nil
}

func (f *Replace) Next() gxui.Focusable {
	next := <-f.input
	switch next {
	case f.find:
		f.status.SetText("Find:")
	case f.replace:
		f.status.SetText("Replace:")
	}
	return next
}
