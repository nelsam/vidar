// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"
	"os"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
)

type Quit struct {
}

func (q Quit) Name() string {
	return "quit"
}

func (q Quit) Menu() string {
	return "File"
}

func (q Quit) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyQ,
	}}
}

func (q Quit) Exec(interface{}) bind.Status {
	// TODO: ask for confirmation if there are changes
	os.Exit(0)
	return bind.Done
}
