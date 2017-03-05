// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package plugins

import (
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
)

func Bindables(theme *basic.Theme) []commander.Bindable {
	return []commander.Bindable{
		GolangHook{Theme: theme},
	}
}
