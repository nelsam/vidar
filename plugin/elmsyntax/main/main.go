// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
)

type ElmHook struct {
}

func (h ElmHook) Name() string {
	return "elm-hook"
}

func (h ElmHook) OpName() string {
	return "focus-location"
}

func (h ElmHook) FileBindables(path string) []bind.Bindable {
	if !strings.HasSuffix(path, ".elm") {
		return nil
	}
	return []bind.Bindable{
		elmsyntax.New(),
	}
}

// Bindables is the main entry point to the command.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	return []bind.Bindable{
		ElmHook{},
	}
}
