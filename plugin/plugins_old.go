// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !linux !go1.8

package plugin

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/commander/bind"
)

func Bindables(cmdr *commander.Commander, driver gxui.Driver, theme *basic.Theme) []bind.Bindable {
	return []bind.Bindable{
		GolangHook{Theme: theme},
	}
}
