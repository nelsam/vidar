// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
)

type FileHook struct {
	Theme *basic.Theme
}

func (h FileHook) Name() string {
	return "edit-hook"
}

func (h FileHook) OpName() string {
	return "open-file"
}

func (h FileHook) FileBindables(string) []bind.Bindable {
	return []bind.Bindable{
		NewSave(h.Theme),
		NewSaveAll(h.Theme),
		NewCloseTab(),
	}
}
