package commands

import (
	"fmt"

	"github.com/nelsam/gxui"
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

func (s *Split) Start(gxui.Control) gxui.Control {
	return nil
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

func (s *Split) Next() gxui.Focusable {
	return nil
}

func (s *Split) Exec(target interface{}) (executed, consume bool) {
	splitter, ok := target.(Splitter)
	if !ok {
		return false, false
	}
	splitter.Split(s.orientation)
	return true, true
}
