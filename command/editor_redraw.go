// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import "github.com/nelsam/vidar/input"

type RedrawableEditor interface {
	input.Editor
	Redraw()
	DataChanged(recreateControls bool)
}

type EditorRedraw struct{}

func (EditorRedraw) Name() string {
	return "editor-redraw"
}

func (EditorRedraw) OpName() string {
	return "input-handler"
}

func (EditorRedraw) Applied(e input.Editor, edits []input.Edit) {
	r := e.(RedrawableEditor)
	r.Redraw()
	r.DataChanged(false)
}
