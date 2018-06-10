// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package ui

// Color represents a color in the UI.
type Color struct {
	R, B, G, A float64
}

// Foregrounder represents a type that has a foreground color.
type Foregrounder interface {
	Foreground() Color
	SetForeground(Color)
}

// Creator represents a type which can create UI elements.
type Creator interface {
	Wrapper

	// Label creates a label in the UI with the passed in string
	// as its text.
	Label(string) Element
}
