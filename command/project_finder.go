// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/setting"
)

type ProjectFinder interface {
	CurrentEditor() *editor.CodeEditor
	Project() settings.Project
}
