// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package caret

import (
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
)

type Commander interface {
	Bindable(name string) bind.Bindable
	Execute(bind.Bindable)
}

type OnEdit struct {
	Commander Commander
}

func (o *OnEdit) Name() string {
	return "caret-movement-on-edit"
}

func (o *OnEdit) OpName() string {
	return "input-handler"
}

func (o *OnEdit) Applied(e text.Editor, edits []text.Edit) {
	h := e.(CaretHandler)
	carets := h.Carets()
	for _, e := range edits {
		carets = o.moveCarets(carets, e)
	}
	m := o.Commander.Bindable("caret-movement").(*Mover)
	o.Commander.Execute(m.To(carets...))
}

func (o *OnEdit) moveCarets(carets []int, e text.Edit) []int {
	delta := len(e.New) - len(e.Old)
	if delta == 0 {
		return carets
	}
	for i, c := range carets {
		if c < e.At {
			continue
		}
		c += delta
		if c < e.At {
			c = e.At
		}
		carets[i] = c
	}
	return carets
}
