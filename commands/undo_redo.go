// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
)

type Undoer interface {
	Undo() (gxui.TextBoxEdit, bool)
}

// An Undo is a command which undoes an action.
type Undo struct {
	status.General

	applier Applier
	undoer  Undoer
	editor  input.Editor
}

func NewUndo(theme gxui.Theme) *Undo {
	u := &Undo{}
	u.Theme = theme
	return u
}

func (u *Undo) Name() string {
	return "undo-last-edit"
}

func (u *Undo) Menu() string {
	return "Edit"
}

func (u *Undo) Reset() {
	u.applier = nil
	u.undoer = nil
	u.editor = nil
}

func (u *Undo) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Undoer:
		u.undoer = src
	case Applier:
		u.applier = src
	case input.Editor:
		u.editor = src
	}
	if u.applier != nil && u.undoer != nil && u.editor != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (u *Undo) Exec() error {
	edit, ok := u.undoer.Undo()
	if !ok {
		u.Warn = "undo: nothing to undo"
		return nil
	}
	u.applier.Apply(u.editor, input.Edit{
		At:  edit.At,
		Old: edit.New,
		New: edit.Old,
	})
	return nil
}

type Redoer interface {
	RedoCurrent() (gxui.TextBoxEdit, bool)
}

// A Redo is a command which redoes an action.
type Redo struct {
	status.General

	editor  input.Editor
	redoer  Redoer
	applier Applier
}

func NewRedo(theme gxui.Theme) *Redo {
	r := &Redo{}
	r.Theme = theme
	return r
}

func (r *Redo) Name() string {
	return "redo-next-edit"
}

func (r *Redo) Menu() string {
	return "Edit"
}

func (r *Redo) Reset() {
	r.editor = nil
	r.redoer = nil
	r.applier = nil
}

func (r *Redo) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Redoer:
		r.redoer = src
	case Applier:
		r.applier = src
	case input.Editor:
		r.editor = src
	}
	if r.applier != nil && r.redoer != nil && r.editor != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (r *Redo) Exec() error {
	edit, ok := r.redoer.RedoCurrent()
	if !ok {
		r.Warn = "redo: nothing to redo"
		return nil
	}
	r.applier.Apply(r.editor, input.Edit{
		At:  edit.At,
		Old: edit.Old,
		New: edit.New,
	})
	return nil
}
