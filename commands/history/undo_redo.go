// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history

import (
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
)

type Applier interface {
	Apply(input.Editor, ...input.Edit)
}

// An Undo is a command which undoes an action.
type Undo struct {
	status.General
	history *History

	applier Applier
	editor  input.Editor
}

func (u *Undo) Name() string {
	return "undo-last-edit"
}

func (u *Undo) Menu() string {
	return "Edit"
}

func (u *Undo) Reset() {
	u.applier = nil
	u.editor = nil
}

func (u *Undo) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Applier:
		u.applier = src
	case input.Editor:
		u.editor = src
	}
	if u.applier != nil && u.editor != nil {
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

	editor  input.Editor
	applier Applier
}

func (r *Redo) Name() string {
	return "redo-next-edit"
}

func (r *Redo) Menu() string {
	return "Edit"
}

func (r *Redo) Reset() {
	r.editor = nil
	r.applier = nil
}

func (r *Redo) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Applier:
		r.applier = src
	case input.Editor:
		r.editor = src
	}
	if r.applier != nil && r.editor != nil {
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
