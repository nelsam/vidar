// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
)

// Commands returns all known commands, in the order they should be
// added to the menu.
func Commands(driver gxui.Driver, theme *basic.Theme, projPane gxui.Control) []commander.Command {
	return []commander.Command{
		// File menu
		NewProjectAdder(driver, theme),
		NewProjectOpener(theme, projPane),
		NewFileOpener(driver, theme),
		NewSelectAll(),
		NewMulti(theme, "File", NewGoImports(theme), NewSave(theme)),
		NewCloseTab(),

		// Edit menu
		NewUndo(),
		NewRedo(theme),
		NewFind(driver, theme),
		NewCopy(driver),
		NewCut(driver),
		NewPaste(driver, theme),
		NewShowSuggestions(),
		NewGotoLine(theme),
		NewGotoDef(theme),
		NewLicenseHeaderUpdate(theme),
		NewGoImports(theme),
		NewComments(),

		// View menu
		NewHorizontalSplit(),
		NewVerticalSplit(),
		NewNextTab(),
		NewPrevTab(),

		// Navigation menu
		NewPrevLine(),
		NewSelectPrevLine(),
		NewNextLine(),
		NewSelectNextLine(),
		NewPrevChar(),
		NewPrevWord(),
		NewSelectPrevChar(),
		NewSelectPrevWord(),
		NewNextChar(),
		NewNextWord(),
		NewSelectNextChar(),
		NewSelectNextWord(),
		NewLineStart(),
		NewSelectLineStart(),
		NewLineEnd(),
		NewSelectLineEnd(),
	}
}
