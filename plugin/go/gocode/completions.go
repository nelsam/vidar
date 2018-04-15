// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gocode

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/status"
)

type TextController interface {
	TextRunes() []rune
	Carets() []int
}

type Completions struct {
	status.General
	gocode *GoCode

	ctrl      TextController
	editor    Editor
	projecter Projecter
	applier   Applier
}

func (c *Completions) Name() string {
	return "show-suggestions"
}

func (c *Completions) Menu() string {
	return "Golang"
}

func (c *Completions) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeySpace,
	}}
}

func (c *Completions) Reset() {
	c.editor = nil
	c.ctrl = nil
	c.projecter = nil
	c.applier = nil
}

func (c *Completions) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case TextController:
		c.ctrl = src
	case Editor:
		c.editor = src
	case Projecter:
		c.projecter = src
	case Applier:
		c.applier = src
	}
	if c.editor != nil && c.ctrl != nil && c.projecter != nil && c.applier != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (c *Completions) Exec() error {
	carets := c.ctrl.Carets()
	if len(carets) != 1 {
		c.Err = "You appear to have multiple carets, but we can only show suggestions for a single caret."
		return errors.New("completions: cannot show suggestions for multiple carets")
	}
	l := newSuggestionList(c.gocode.driver, c.Theme.(*basic.Theme), c.projecter.Project(), c.editor, c.ctrl, c.applier)
	c.gocode.set(c.editor, l, carets[0])
	return nil
}
