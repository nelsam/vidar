// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/setting"
)

type ProjectFinder interface {
	CurrentEditor() text.Editor
	Project() setting.Project
}
