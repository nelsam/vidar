// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package core

import (
	"github.com/nelsam/vidar/core/gxui"
	"github.com/nelsam/vidar/ui"
)

func UIs() []ui.Creator {
	return []ui.Creator{gxui.Creator{}}
}
