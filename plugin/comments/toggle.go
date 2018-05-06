// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package comments

import (
	"fmt"
	"regexp"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
)

type Applier interface {
	Apply(input.Editor, ...input.Edit)
}

type Editor interface {
	input.Editor
}

type Selecter interface {
	Selections() []gxui.TextSelection
}

type Toggle struct {
	editor   input.Editor
	applier  Applier
	selecter Selecter
}

func NewToggle() *Toggle {
	return &Toggle{}
}

func (c *Toggle) Name() string {
	return "toggle-comments"
}

func (c *Toggle) Menu() string {
	return "Golang"
}

func (c *Toggle) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeySlash,
	}}
}

func (t *Toggle) Reset() {
	t.editor = nil
	t.applier = nil
	t.selecter = nil
}

func (t *Toggle) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Applier:
		t.applier = src
	case input.Editor:
		t.editor = src
	case Selecter:
		t.selecter = src
	}
	if t.editor != nil && t.applier != nil && t.selecter != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (t *Toggle) Exec() error {
	selections := t.selecter.Selections()

	var edits []input.Edit
	for i := len(selections) - 1; i >= 0; i-- {
		begin, end := selections[i].Start(), selections[i].End()
		str := t.editor.Text()[begin:end]
		re, replace := regexpReplace(str)
		newstr := re.ReplaceAllString(str, replace)

		edits = append(edits, input.Edit{
			At:  int(begin),
			Old: []rune(str),
			New: []rune(newstr),
		})
	}
	t.applier.Apply(t.editor, edits...)
	return nil
}

func regexpReplace(str string) (*regexp.Regexp, string) {
	if regexp.MustCompile(`^(\s*?)//`).MatchString(str) {
		return regexp.MustCompile(`(?m)^(\s*?)//(.*)$`), `${1}${2}`
	}
	return regexp.MustCompile("(?m)^(.*)$"), `//${1}`
}
