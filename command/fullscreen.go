// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"
	"log"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
)

type Fullscreener interface {
	Fullscreen() bool
	SetFullscreen(bool)
}

type Fullscreen struct {
}

func (f Fullscreen) Name() string {
	return "toggle-fullscreen"
}

func (f Fullscreen) Menu() string {
	return "View"
}

func (f Fullscreen) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Key: gxui.KeyF11,
	}}
}

func (f Fullscreen) Exec(e interface{}) bind.Status {
	fs, ok := e.(Fullscreener)
	if !ok {
		log.Printf("Type %T is not a fullscreener", e)
		return bind.Waiting
	}
	fs.SetFullscreen(!fs.Fullscreen())
	return bind.Done
}
