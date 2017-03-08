// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package command contains types that plugins use to access
// functionality of the commander.Commander.  It is kept separate
// from the plugin package so that plugins can import it without
// risking rapid changes requiring rebuilds.
//
// For more information, see vidar's plugin package documentation.
package command

import "github.com/nelsam/vidar/commander/bind"

// A Commander is a type that can look up bind.Commands by name.
type Commander interface {
	Command(name string) bind.Command
}
