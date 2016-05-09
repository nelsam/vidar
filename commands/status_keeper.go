// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander"
)

// statusKeeper is a type used to keep track of a command's status.
// It implements commander.Statuser so that other types can simply
// embed it and rely on its functionality.
//
// Types that embed this *must* set its theme field.
type statusKeeper struct {
	theme gxui.Theme

	err  string
	warn string
	info string
}

func (s *statusKeeper) message() (gxui.Color, string) {
	if s.err != "" {
		return commander.ColorErr, s.err
	}
	if s.warn != "" {
		return commander.ColorWarn, s.warn
	}
	return commander.ColorInfo, s.info
}

func (s *statusKeeper) clear() {
	s.err = ""
	s.warn = ""
	s.info = ""
}

func (s *statusKeeper) Status() gxui.Control {
	defer s.clear()
	color, message := s.message()
	if message == "" {
		return nil
	}
	label := s.theme.CreateLabel()
	label.SetColor(color)
	label.SetText(message)
	return label
}
