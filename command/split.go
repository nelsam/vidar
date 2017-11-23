// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
)

type Splitter interface {
	Split(gxui.Orientation)
}

type Split struct {
	orientation gxui.Orientation
}

func NewHorizontalSplit() *Split {
	return &Split{
		orientation: gxui.Horizontal,
	}
}

func NewVerticalSplit() *Split {
	return &Split{
		orientation: gxui.Vertical,
	}
}

func (s *Split) Name() string {
	switch s.orientation {
	case gxui.Horizontal:
		return "split-view-horizontally"
	case gxui.Vertical:
		return "split-view-vertically"
	default:
		panic(fmt.Errorf("Orientation %d is invalid", s.orientation))
	}
}

func (s *Split) Menu() string {
	return "View"
}

func (s *Split) Exec(target interface{}) bind.Status {
	// TODO: refocus the currently focused editor after we split.
	splitter, ok := target.(Splitter)
	if !ok {
		return bind.Waiting
	}
	splitter.Split(s.orientation)
	return bind.Done
}
