// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import "github.com/nelsam/gxui"

// GenericStatuser is a type used to keep track of a generic status.
// It implements Statuser so that other types can simply embed it
// and rely on its functionality.
//
// Types that embed this *must* set its Theme field.
type GenericStatuser struct {
	Theme gxui.Theme

	Err  string
	Warn string
	Info string
}

func (s *GenericStatuser) message() (gxui.Color, string) {
	if s.Err != "" {
		return ColorErr, s.Err
	}
	if s.Warn != "" {
		return ColorWarn, s.Warn
	}
	return ColorInfo, s.Info
}

// Clear clears all fields in s.
func (s *GenericStatuser) Clear() {
	s.Err = ""
	s.Warn = ""
	s.Info = ""
}

// Status returns the gxui.Control element that displays s's status.
func (s *GenericStatuser) Status() gxui.Control {
	defer s.Clear()
	color, message := s.message()
	if message == "" {
		return nil
	}
	label := s.Theme.CreateLabel()
	label.SetColor(color)
	label.SetText(message)
	return label
}
