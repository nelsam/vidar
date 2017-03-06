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
func Commands(driver gxui.Driver, theme *basic.Theme) []commander.Command {
	return []commander.Command{
		NewProjectAdder(driver, theme),
		NewProjectOpener(theme),
		NewFileOpener(driver, theme),
	}
}

// Hooks returns all known hooks that trigger off of events rather
// than key bindings.
func Hooks(driver gxui.Driver, theme *basic.Theme) []commander.CommandHook {
	return []commander.CommandHook{
		FileHook{Theme: theme},
		EditHook{Theme: theme, Driver: driver},
		ViewHook{},
		NavHook{},
	}
}
