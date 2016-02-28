package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
)

// An Undo is a command which undoes an action.
type Undo struct {
}

func NewUndo() *Undo {
	return &Undo{}
}

func (u *Undo) Start(gxui.Control) gxui.Control {
	return nil
}

func (u *Undo) Name() string {
	return "undo-last-edit"
}

func (u *Undo) Next() gxui.Focusable {
	return nil
}

func (u *Undo) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	history := editor.History()
	edit := history.Undo()
	if edit.Delta < 0 {
		edit.At += 1
	}
	text, _ := editor.Controller().ReplaceAt(editor.Runes(), edit.At, edit.At+len(edit.New), edit.Old)
	editor.Controller().SetTextRunes(text)
	return true, true
}

// A Redo is a command which redoes an action.
type Redo struct {
	theme  gxui.Theme
	editor *editor.CodeEditor
}

func NewRedo(theme gxui.Theme) *Redo {
	return &Redo{
		theme: theme,
	}
}

func (r *Redo) Start(target gxui.Control) gxui.Control {
	r.editor = findEditor(target)
	return nil
}

func (r *Redo) Name() string {
	return "redo-next-edit"
}

func (r *Redo) Next() gxui.Focusable {
	return nil
}

func (r *Redo) Exec(interface{}) (executed, consume bool) {
	history := r.editor.History()
	edit := history.Redo(0)
	text, _ := r.editor.Controller().ReplaceAt(r.editor.Runes(), edit.At, edit.At+len(edit.Old), edit.New)
	r.editor.Controller().SetTextRunes(text)
	return true, true
}
