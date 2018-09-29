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

type OnMove struct {
	Commander Commander
}

func (o *OnMove) Name() string {
	return "scroll-on-caret-movement"
}

func (o *OnMove) OpName() string {
	return "caret-movement"
}

func (o *OnMove) Moved(e input.Editor, carets []int) {
	s := o.Commander.Bindable("scroll").(*Scroller)
	o.Commander.Execute(s.For(ToRune(carets[0])))
}
