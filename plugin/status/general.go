// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// Package status contains types that plugins can make use of in
// their bindables for displaying status to the user.  This package
// is kept on its own to reduce the frequency with which plugins
// importing it must be rebuilt.
//
// For more information, see vidar's plugin package documentation.
package status

import "github.com/nelsam/gxui"

var (
	ColorErr = gxui.Color{
		R: 1,
		G: 0.2,
		B: 0,
		A: 1,
	}
	ColorWarn = gxui.Color{
		R: 0.8,
		G: 0.7,
		B: 0.1,
		A: 1,
	}
	ColorInfo = gxui.Color{
		R: 0.1,
		G: 0.8,
		B: 0,
		A: 1,
	}
)

// General is a type used to keep track of a general status.
// It implements commander.Statuser so that other types can
// simply embed it and rely on its functionality.
//
// Types that embed this *must* set its Theme field.
type General struct {
	Theme gxui.Theme

	Err  string
	Warn string
	Info string
}

func (s *General) message() (gxui.Color, string) {
	if s.Err != "" {
		return ColorErr, s.Err
	}
	if s.Warn != "" {
		return ColorWarn, s.Warn
	}
	return ColorInfo, s.Info
}

// Clear clears all fields in s.
func (s *General) Clear() {
	s.Err = ""
	s.Warn = ""
	s.Info = ""
}

// Status returns the gxui.Control element that displays s's status.
func (s *General) Status() gxui.Control {
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
