// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package scroll

import (
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
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
	if len(edits) == 1 && len(edits[0].New) == 0 {
		s := o.Commander.Bindable("scroll-setting").(*Scroller)
		o.Commander.Execute(s.To(ToRune, edits[0].At))
	}
}
