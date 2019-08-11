// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !linux,!darwin linux,!go1.8 darwin,!go1.10

package plugin

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/bind"
)

func Bindables(cmdr *commander.Commander, driver gxui.Driver, theme *basic.Theme) []bind.Bindable {
	return []bind.Bindable{
		GolangHook{Theme: theme, Driver: driver},
	}
}
