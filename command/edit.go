// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/bind"
)

type EditHook struct {
	Driver gxui.Driver
	Theme  *basic.Theme
}

func (h EditHook) Name() string {
	return "edit-hook"
}

func (h EditHook) OpName() string {
	return "focus-location"
}

func (h EditHook) FileBindables(string) []bind.Bindable {
	return []bind.Bindable{
		NewSelectAll(),
		NewFind(h.Driver, h.Theme),
		NewRegexFind(h.Driver, h.Theme),
		NewReplace(h.Driver, h.Theme),
		NewCopy(h.Driver),
		NewCut(h.Driver),
		NewPaste(h.Driver, h.Theme),
		NewGotoLine(h.Theme),
	}
}
