// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"regexp"

	"github.com/nelsam/gxui"
)

type Comments struct {
}

func NewComments() *Comments {
	return &Comments{}
}

func (c *Comments) Name() string {
	return "toggle-comments"
}

func (c *Comments) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return true, true
	}

	selections := editor.Controller().Selections()

	for i := selections.Len(); i != 0; i-- {
		begin, end := selections.GetInterval(i - 1)
		str := string(editor.Runes()[begin:end])
		re, replace := getRegexpAndReplace(str)
		newstr := re.ReplaceAllString(str, replace)
		newRunes, edit := editor.Controller().ReplaceAt(
			editor.Runes(), int(begin), int(end), []rune(newstr))

		editor.Controller().SetTextEdits(newRunes, []gxui.TextBoxEdit{edit})

	}
	return true, true

}

func getRegexpAndReplace(str string) (*regexp.Regexp, string) {

	if regexp.MustCompile(`^(\s*?)//`).MatchString(str) {
		return regexp.MustCompile(`(?m)^(\s*?)//(.*)$`), `${1}${2}`
	} else {
		return regexp.MustCompile("(?m)^(.*)$"), `//${1}`
	}
}
