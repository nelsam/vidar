// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
)

// An Undo is a command which undoes an action.
type Undo struct {
	statusKeeper
}

func NewUndo(theme gxui.Theme) *Undo {
	return &Undo{statusKeeper: statusKeeper{theme: theme}}
}

func (u *Undo) Name() string {
	return "undo-last-edit"
}

func (u *Undo) Menu() string {
	return "Edit"
}

func (u *Undo) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		u.err = "undo: no file open"
		return true, true
	}
	history := editor.History()
	edit, ok := history.Undo()
	if !ok {
		u.warn = "undo: nothing to undo"
		return true, true
	}
	text, _ := editor.Controller().ReplaceAt(editor.Runes(), edit.At, edit.At+len(edit.New), edit.Old)
	editor.Controller().SetTextRunes(text)
	absDelta := edit.Delta
	if absDelta < 0 {
		absDelta = -absDelta
	}
	editor.Controller().SetSelection(gxui.CreateTextSelection(edit.At, edit.At+absDelta, true))
	editor.ScrollToRune(edit.At)
	return true, true
}

// A Redo is a command which redoes an action.
type Redo struct {
	statusKeeper

	theme  gxui.Theme
	editor *editor.CodeEditor
}

func NewRedo(theme gxui.Theme) *Redo {
	return &Redo{
		theme:        theme,
		statusKeeper: statusKeeper{theme: theme},
	}
}

func (r *Redo) Start(target gxui.Control) gxui.Control {
	r.editor = findEditor(target)
	return nil
}

func (r *Redo) Name() string {
	return "redo-next-edit"
}

func (r *Redo) Menu() string {
	return "Edit"
}

func (r *Redo) Next() gxui.Focusable {
	return nil
}

func (r *Redo) Exec(interface{}) (executed, consume bool) {
	if r.editor == nil {
		r.err = "redo: no file open"
		return true, true
	}
	history := r.editor.History()
	edit, ok := history.RedoCurrent()
	if !ok {
		r.warn = "redo: nothing to redo"
		return true, true
	}
	text, _ := r.editor.Controller().ReplaceAt(r.editor.Runes(), edit.At, edit.At+len(edit.Old), edit.New)
	r.editor.Controller().SetTextRunes(text)
	absDelta := edit.Delta
	if absDelta < 0 {
		absDelta = -absDelta
	}
	r.editor.Controller().SetSelection(gxui.CreateTextSelection(edit.At, edit.At+absDelta, true))
	r.editor.ScrollToRune(edit.At)
	return true, true
}
