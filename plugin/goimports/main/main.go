// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
	"github.com/nelsam/vidar/plugin/goimports"
)

type GolangHook struct {
	Theme gxui.Theme
}

func (h GolangHook) Name() string {
	return "golang-hook"
}

func (h GolangHook) OpName() string {
	return "open-file"
}

func (h GolangHook) FileBindables(path string) []bind.Bindable {
	if !strings.HasSuffix(path, ".go") {
		return nil
	}
	return []bind.Bindable{
		goimports.New(h.Theme),
		goimports.OnSave{},
	}
}

// Bindables is the main entry point to the command.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	return []bind.Bindable{
		GolangHook{Theme: theme},
	}
}
