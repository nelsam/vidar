// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
)

type RegexFind struct {
	finder *Find
}

func NewRegexFind(driver gxui.Driver, theme *basic.Theme) *RegexFind {
	f := &RegexFind{}
	f.finder = NewFind(driver, theme)
	return f
}

func (f *RegexFind) Start(control gxui.Control) gxui.Control {
	f.finder.editor = findEditor(control)
	if f.finder.editor == nil {
		return nil
	}
	f.finder.pattern = newFindBox(f.finder.driver, f.finder.theme)
	f.finder.AddChild(f.finder.pattern)
	f.finder.pattern.OnTextChanged(func([]gxui.TextBoxEdit) {
		f.finder.editor.Controller().ClearSelections()
		needle := f.finder.pattern.Text()
		if len(needle) == 0 {
			f.finder.display.SetText("Start typing to search")
			return
		}
		exp, err := regexp.Compile(needle)
		if err != nil {
			f.finder.display.SetText("Incorrect regexp")
			return
		}
		f.finder.selections = []int{}
		haystack := f.finder.editor.Text()
		arr := exp.FindAllStringIndex(haystack, -1)
		if arr == nil {
			f.finder.display.SetText("Match not found")
			return
		}

		var selections []gxui.TextSelection

		for _, indexes := range arr {
			for i := 0; i < len(indexes); i += 2 {
				start := utf8.RuneCountInString(haystack[0:indexes[i]])
				end := start + utf8.RuneCountInString(haystack[indexes[i]:indexes[i+1]])
				selection := gxui.CreateTextSelection(start, end, false)
				selections = append(selections, selection)
				f.finder.selections = append(f.finder.selections, start)
			}
		}
		f.finder.selection = len(selections) - 1
		f.finder.editor.SelectSlice(selections)
		f.finder.display.SetText(fmt.Sprintf("%s: %d results found", needle, len(selections)))
	})
	return f.finder.display
}

func (f *RegexFind) Name() string {
	return "regex-find"
}

func (f *RegexFind) Menu() string {
	return "Edit"
}

func (f *RegexFind) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModAlt,
		Key:      gxui.KeyF,
	}}
}

func (f *RegexFind) Next() gxui.Focusable {
	return f.finder
}
