// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build linux,go1.8

// Package plugin contains logic for registering and binding plugins
// with vidar.  While designing plugins, we noticed that any change
// to the functionality in a common import between vidar and the
// plugin would cause the plugin to fail to load.
//
// As a result, you'll see several very small sub-packages containing
// only interface types that are used either as parameters or return
// types for methods.  This allows plugins to import those small
// sub-packages, which change far less frequently than the main
// packages.
//
// When you see one of these small sub-packages, consider them
// packages that plugins should feel comfortable importing with a
// very low risk of needing to be rebuilt for new versions of vidar.
//
// A plugin must export a single function, named Bindables, which
// returns all bind.Bindables that the plugin provides.  It must
// accept arguments of type command.Commander, gxui.Driver, and
// gxui.Theme, and its return type must be []bind.Bindable.
package plugin

import (
	"log"
	"plugin"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
	"github.com/nelsam/vidar/settings"
)

const lookupName = "Bindables"

// Bindables returns all bindables that are found via plugins
// in the plugin directory.
//
// For each plugin, Bindables will look up a Bindables function
// and expect it to accept a Commander, a gxui.Driver, and a
// gxui.Theme as arguments.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	var bindables []bind.Bindable
	for _, path := range settings.Plugins() {
		plugin, err := plugin.Open(path)
		if err != nil {
			log.Printf("Error opening plugin at %s: %s", path, err)
			continue
		}
		c, err := plugin.Lookup(lookupName)
		if err != nil {
			log.Printf("Error looking up constructor %s in plugin %s: %s", lookupName, path, err)
			continue
		}

		var newBindables []bind.Bindable
		switch construct := c.(type) {
		case func(command.Commander, gxui.Driver, gxui.Theme) []bind.Bindable:
			newBindables = construct(cmdr, driver, theme)
		default:
			log.Printf("Error: don't know how to call constructor of type %T from plugin %s", c, path)
			continue
		}
		bindables = append(bindables, newBindables...)
	}
	return bindables
}
