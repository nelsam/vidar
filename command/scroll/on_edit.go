// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package scroll

import (
	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/input"
)

type Commander interface {
	Bindable(name string) bind.Bindable
	Execute(bind.Bindable)
}

type OnEdit struct {
	Commander Commander
}

func (o *OnEdit) Name() string {
	return "scroll-movement-on-edit"
}

func (o *OnEdit) OpName() string {
	return "input-handler"
}

func (o *OnEdit) Applied(e input.Editor, edits []input.Edit) {
	// TODO: handling multiple editing
	if len(edits) == 1 {
		edit := edits[0]
		s := o.Commander.Bindable("scroll").(*Scroller)
		if len(edit.New) == 0 {
			if len(edit.Old) == 1 {
				o.Commander.Execute(s.For(ToRune(edit.At, Up)))
			} else {
				o.Commander.Execute(s.For(ToOldOffset()))
			}
		} else if len(edit.Old) == 0 {
			o.Commander.Execute(s.For(ToRune(edit.At+len(edit.New), NoDirection)))
		}
	}
}
