// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
	"github.com/nelsam/vidar/plugin/vfmt"
)

type VlangHook struct {
	Theme gxui.Theme
}

func (h VlangHook) Name() string {
	return "vlang-hook"
}

func (h VlangHook) OpName() string {
	return "focus-location"
}

func (h VlangHook) FileBindables(path string) []bind.Bindable {
	if !strings.HasSuffix(path, ".v") {
		return nil
	}
	return []bind.Bindable{
		vfmt.New(h.Theme),
		vfmt.OnSave{},
	}
}

// Bindables is the main entry point to the command.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	return []bind.Bindable{
		VlangHook{Theme: theme},
	}
}
