package commands

import (
	"log"
	"regexp"

	"github.com/nelsam/gxui"
)

type Comments struct {
	driver gxui.Driver
}

func NewComments(driver gxui.Driver) *Comments {
	return &Comments{
		driver: driver,
	}
}

func (c *Comments) Name() string {
	return "toggle-comments"
}

func (c *Comments) Start(gxui.Control) gxui.Control {
	return nil
}

func (c *Comments) Next() gxui.Focusable {
	return nil
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
	if editor.Controller().SelectionCount() != 1 {
		log.Print("Warning: toggle comments can currently only comments the first selection")
	}

	selection := editor.Controller().FirstSelection()
	begin, end := selection.Start(), selection.End()
	str := string(editor.Runes()[begin:end])
	re, replace := getRegexpAndReplace(str)
	newstr := re.ReplaceAllString(str, replace)
	newRunes, edit := editor.Controller().ReplaceAt(
		editor.Runes(), begin, end, []rune(newstr))

	editor.Controller().SetTextEdits(newRunes, []gxui.TextBoxEdit{edit})

	return true, true

}

func getRegexpAndReplace(str string) (*regexp.Regexp, string) {

	if regexp.MustCompile(`^(\s*?)//`).MatchString(str) {
		return regexp.MustCompile(`(?m)^(\s*?)//(.*)$`), `${1}${2}`
	} else {
		return regexp.MustCompile("(?m)^(.*)$"), `//${1}`
	}
}
