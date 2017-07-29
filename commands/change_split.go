// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/editor"
)

type SplitShifter interface {
	ShiftSplit(editor.Direction)
}

type ChangeFocus struct {
	direction editor.Direction
	name      string
}

func NewFocusRight() *ChangeFocus {
	return &ChangeFocus{direction: editor.Right, name: "focus-right"}
}

func NewFocusLeft() *ChangeFocus {
	return &ChangeFocus{direction: editor.Left, name: "focus-left"}
}

func NewFocusUp() *ChangeFocus {
	return &ChangeFocus{direction: editor.Up, name: "focus-up"}
}

func NewFocusDown() *ChangeFocus {
	return &ChangeFocus{direction: editor.Down, name: "focus-down"}
}

func (p *ChangeFocus) Name() string {
	return p.name
}

func (p *ChangeFocus) Menu() string {
	return "View"
}

func (p *ChangeFocus) Exec(target interface{}) bind.Status {
	shifter, ok := target.(SplitShifter)
	if !ok {
		return bind.Waiting
	}
	shifter.ShiftSplit(p.direction)
	return bind.Done
}
