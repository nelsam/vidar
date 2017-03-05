// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
)

type EditHook struct {
	Driver gxui.Driver
	Theme  *basic.Theme
}

func (h EditHook) Name() string {
	return "edit-hook"
}

func (h EditHook) CommandName() string {
	return "open-file"
}

func (h EditHook) FileBindables(string) []commander.Bindable {
	return []commander.Bindable{
		NewUndo(h.Theme),
		NewRedo(h.Theme),
		NewSelectAll(),
		NewFind(h.Driver, h.Theme),
		NewRegexFind(h.Driver, h.Theme),
		NewCopy(h.Driver),
		NewCut(h.Driver),
		NewPaste(h.Driver, h.Theme),
		NewShowSuggestions(),
		NewGotoLine(h.Theme),
	}
}
