// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
	"github.com/nelsam/vidar/plugin/vsyntax"
)

type vlangHook struct{}

func (h vlangHook) Name() string {
	return "vlang-hook"
}

func (h vlangHook) OpName() string {
	return "focus-location"
}

func (h vlangHook) FileBindables(path string) []bind.Bindable {
	if !strings.HasSuffix(path, ".v") {
		return nil
	}
	return []bind.Bindable{
		vsyntax.New(),
	}
}

// Bindables is the main entry point to the plugin.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	return []bind.Bindable{vlangHook{}}
}
