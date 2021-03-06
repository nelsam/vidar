// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/command"
	"github.com/nelsam/vidar/plugin/license"
)

type LicenseHook struct {
	Theme gxui.Theme
}

func (h LicenseHook) Name() string {
	return "license-lang-hook"
}

func (h LicenseHook) OpName() string {
	return "focus-location"
}

func (h LicenseHook) FileBindables(path string) []bind.Bindable {
	switch {
	case strings.HasSuffix(path, ".go"), strings.HasSuffix(path, ".v"):
		return []bind.Bindable{
			license.NewHeaderUpdate(h.Theme),
		}
	default:
		return nil
	}
}

// Bindables is the main entry point to the command.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	return []bind.Bindable{
		LicenseHook{Theme: theme},
	}
}
