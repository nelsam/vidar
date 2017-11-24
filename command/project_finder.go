// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/setting"
)

type ProjectFinder interface {
	CurrentEditor() input.Editor
	Project() setting.Project
}
