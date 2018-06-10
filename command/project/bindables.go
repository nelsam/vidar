// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package project

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/bind"
)

// Bindables returns all bindables that relate to projects.
func Bindables(driver gxui.Driver, theme *basic.Theme) []bind.Bindable {
	return []bind.Bindable{
		&Open{},
		NewAdd(driver, theme),
		NewFind(theme),
	}
}
