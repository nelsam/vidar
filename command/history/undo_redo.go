// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/plugin/status"
)

type Applier interface {
	Apply(text.Editor, ...text.Edit)
}

type OnOpen struct {
	theme *basic.Theme
}

func (o OnOpen) Name() string {
	return "undo-redo-open-hook"
}

func (o OnOpen) OpName() string {
	return "focus-location"
}

func (o OnOpen) FileBindables(string) []bind.Bindable {
	u := &Undo{}
	u.Theme = o.theme
	r := &Redo{}
	r.Theme = o.theme
	return []bind.Bindable{u, r}
}

// An Undo is a command which undoes an action.
type Undo struct {
	status.General

	history *History
	applier Applier
	editor  text.Editor
}

func (u *Undo) Name() string {
	return "undo-last-edit"
}

func (u *Undo) Menu() string {
	return "Edit"
}

func (u *Undo) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyZ,
	}}
}

func (u *Undo) Reset() {
	u.history = nil
	u.applier = nil
	u.editor = nil
}

func (u *Undo) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case *History:
		u.history = src
	case Applier:
		u.applier = src
	case text.Editor:
		u.editor = src
	}
	if u.applier != nil && u.editor != nil && u.history != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (u *Undo) Exec() error {
	edit := u.history.Rewind()
	if edit.At == -1 {
		u.Warn = "undo: nothing to undo"
		return nil
	}
	u.applier.Apply(u.editor, edit)
	return nil
}

// A Redo is a command which redoes an action.
type Redo struct {
	status.General

	history *History
	editor  text.Editor
	applier Applier
}

func (r *Redo) Name() string {
	return "redo-next-edit"
}

func (r *Redo) Menu() string {
	return "Edit"
}

func (r *Redo) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyZ,
	}}
}

func (r *Redo) Reset() {
	r.history = nil
	r.editor = nil
	r.applier = nil
}

func (r *Redo) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case *History:
		r.history = src
	case Applier:
		r.applier = src
	case text.Editor:
		r.editor = src
	}
	if r.applier != nil && r.editor != nil && r.history != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (r *Redo) Exec() error {
	// Overflow will just result in a high number, so no need to
	// check for it.
	edit := r.history.FastForward(r.history.Branches() - 1)
	if edit.At == -1 {
		r.Warn = "redo: nothing to redo"
		return nil
	}
	r.applier.Apply(r.editor, edit)
	return nil
}
