// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package ui

import "github.com/nelsam/vidar/theme"

// Span represents a span of text.
type Span struct {
	Highlight     theme.Highlight
	Start, Length uint64
}

// Offset tells a Layout how far away from the default position
// the element should be placed.
type Offset struct {
	X, Y int
}

// Size represents an element's size.
type Size struct {
	Width, Height uint64
}

// Direction represents a direction in the UI, for things
// like choosing which edge to align elements with or shifting
// elements in a direction.
type Direction int

const (
	// Not all functions/methods will support multiple directions
	// at once (e.g. Up | Right); read the docs to understand how
	// to use these values.
	Up Direction = 1 << iota
	Right
	Down
	Left
)

// Alignment represents a method of aligning differently-sized
// elements.
type Alignment int

const (
	// Start will typically be either top or left, and End the
	// opposite.  However, double-check the docs of the function
	// that you're passing these values in to in case there are
	// deviations.
	Start = iota
	Center
	End
)
