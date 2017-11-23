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

type RegexFind Find

func NewRegexFind(driver gxui.Driver, theme *basic.Theme) *RegexFind {
	finder := &RegexFind{}
	finder.Init(driver, theme)
	return finder
}

func (f *RegexFind) Init(driver gxui.Driver, theme *basic.Theme) {
	f.driver = driver
	f.theme = theme
}

func (f *RegexFind) Start(control gxui.Control) gxui.Control {
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
		exp, err := regexp.Compile(needle)
		if err != nil {
			f.display.SetText("Incorrect regexp")
			return
		}
		haystack := f.editor.Text()
		arr := exp.FindAllStringIndex(haystack, -1)
		if arr == nil {
			f.display.SetText("Match not found")
			return
		}

		var selections gxui.TextSelectionList

		for _, indexes := range arr {
			for i := 0; i < len(indexes); i += 2 {
				start := utf8.RuneCountInString(haystack[0:indexes[i]])
				end := start + utf8.RuneCountInString(haystack[indexes[i]:indexes[i+1]])
				selection := gxui.CreateTextSelection(start, end, false)
				selections = append(selections, selection)

			}
		}
		f.editor.Select(selections)
		f.display.SetText(fmt.Sprintf("%s: %d results found", needle, len(selections)))
	})
	return f.display
}

func (f *RegexFind) Name() string {
	return "regex-find"
}

func (f *RegexFind) Menu() string {
	return "Edit"
}

func (f *RegexFind) Next() gxui.Focusable {
	next := f.next
	f.next = nil
	return next
}
