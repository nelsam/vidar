// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/plugins/comments"
	"github.com/nelsam/vidar/plugins/godef"
	"github.com/nelsam/vidar/plugins/goimports"
	"github.com/nelsam/vidar/plugins/license"
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
		NewMulti(theme, "File", goimports.New(theme), NewSave(theme)),
		NewSaveAll(theme),
		NewCloseTab(),

		// Edit menu
		NewUndo(theme),
		NewRedo(theme),
		NewFind(driver, theme),
		NewRegexFind(driver, theme),
		NewCopy(driver),
		NewCut(driver),
		NewPaste(driver, theme),
		NewShowSuggestions(),
		NewGotoLine(theme),
		godef.New(theme),
		license.NewHeaderUpdate(theme),
		goimports.New(theme),
		comments.NewToggle(),

		// View menu
		NewHorizontalSplit(),
		NewVerticalSplit(),
		NewNextTab(),
		NewPrevTab(),
		NewFocusUp(),
		NewFocusDown(),
		NewFocusLeft(),
		NewFocusRight(),

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
