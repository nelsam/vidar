// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
)

// TODO: command/hook ordering is handled in commander, so we probably
// don't need to split these up any more.

// Commands returns all known commands, in the order they should be
// added to the menu.
func Commands(_ command.Commander, driver gxui.Driver, theme *basic.Theme) []bind.Command {
	return []bind.Command{
		NewProjectAdder(driver, theme),
		NewProjectOpener(theme),
		NewFileOpener(driver, theme),
	}
}

// Hooks returns all known hooks that trigger off of events rather
// than key bindings.
func Hooks(cmdr command.Commander, driver gxui.Driver, theme *basic.Theme) []bind.Bindable {
	return []bind.Bindable{
		FileHook{Theme: theme},
		EditHook{Theme: theme, Driver: driver},
		ViewHook{},
		NavHook{Commander: cmdr},
	}
}
