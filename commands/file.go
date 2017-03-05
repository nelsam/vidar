// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
)

type FileHook struct {
	Theme *basic.Theme
}

func (h FileHook) Name() string {
	return "edit-hook"
}

func (h FileHook) CommandName() string {
	return "open-file"
}

func (h FileHook) FileBindables(string) []commander.Bindable {
	return []commander.Bindable{
		NewSave(h.Theme),
		NewSaveAll(h.Theme),
		NewCloseTab(),
	}
}
