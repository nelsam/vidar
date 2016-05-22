// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
)

func Commands(driver gxui.Driver, theme *basic.Theme, projPane gxui.Control) []commander.Command {
	return []commander.Command{
		NewProjectAdder(driver, theme),
		NewProjectOpener(theme, projPane),
		NewFileOpener(driver, theme),
		NewMulti(theme, "File", NewGoImports(theme), NewSave(theme)),
		NewCloseTab(),

		NewUndo(),
		NewRedo(theme),
		NewFinder(driver, theme),
		NewCopy(driver),
		NewCut(driver),
		NewPaste(driver, theme),
		NewGotoLine(theme),
		NewGotoDef(theme),
		NewLicenseHeaderUpdate(theme),
		NewGoImports(theme),
		NewComments(),

		NewHorizontalSplit(),
		NewVerticalSplit(),
	}
}
