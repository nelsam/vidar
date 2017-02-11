// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package comments

import (
	"regexp"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
)

type EditorContainer interface {
	CurrentEditor() *editor.CodeEditor
}

type Toggle struct {
}

func NewToggle() *Toggle {
	return &Toggle{}
}

func (c *Toggle) Name() string {
	return "toggle-comments"
}

func (c *Toggle) Menu() string {
	return "Edit"
}

func (c *Toggle) Exec(target interface{}) (executed, consume bool) {
	container, ok := target.(EditorContainer)
	if !ok {
		return false, false
	}
	editor := container.CurrentEditor()
	if editor == nil {
		return true, true
	}

	selections := editor.Controller().Selections()

	for i := selections.Len(); i != 0; i-- {
		begin, end := selections.Interval(i - 1)
		str := string(editor.Runes()[begin:end])
		re, replace := regexpReplace(str)
		newstr := re.ReplaceAllString(str, replace)
		newRunes, edit := editor.Controller().ReplaceAt(
			editor.Runes(), int(begin), int(end), []rune(newstr))

		editor.Controller().SetTextEdits(newRunes, []gxui.TextBoxEdit{edit})

	}
	return true, true

}

func regexpReplace(str string) (*regexp.Regexp, string) {
	if regexp.MustCompile(`^(\s*?)//`).MatchString(str) {
		return regexp.MustCompile(`(?m)^(\s*?)//(.*)$`), `${1}${2}`
	}
	return regexp.MustCompile("(?m)^(.*)$"), `//${1}`
}
