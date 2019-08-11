// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package ui

import "github.com/nelsam/vidar/input"

// LayoutCfg is a configuration for options passed to the Add
// method of a layout.  Since we need to support sort of "generic"
// options (rather than relying on library-specific options), we
// need something like this.
//
// UI library wrappers will create an empty config, then pass
// it to the functional options passed in to the Add call.
//
// All fields are pointers so that they default to `nil`.
type LayoutCfg struct {
	Alignment Alignment
	Offset    *Offset
	NextTo    *NextTo
	Name      *string
}

// NextTo is a type which contains an element and a direction,
// giving layouts information that an element should be touching
// another element on the provided side.
type NextTo struct {
	Sibling interface{}
	Side    Direction
}

// LayoutOpt is a function that returns a modified LayoutCfg
type LayoutOpt func(LayoutCfg) LayoutCfg

// LayoutOffset is a LayoutOpt that offsets an element's position.
func LayoutOffset(x, y int) LayoutOpt {
	return func(c LayoutCfg) LayoutCfg {
		if c.Offset == nil {
			c.Offset = &Offset{}
		}
		c.Offset.X += x
		c.Offset.Y += y
		return c
	}
}

// LayoutNextTo is a LayoutOpt that tells the Layout to
// align the element in relation to another element.  This
// should work fine with tab and split layouts, but may not
// work with other layout types, depending on the UI library.
//
// If elem is nil, then the new element will be aligned next
// to _all other elements_, i.e. either first or last.
func LayoutNextTo(elem interface{}, dir Direction) LayoutOpt {
	return func(c LayoutCfg) LayoutCfg {
		c.NextTo = &NextTo{
			Sibling: elem,
			Side:    dir,
		}
		return c
	}
}

// LayoutName is a LayoutOpt that tells the Layout the name
// of the new element.  This is most important for tabs, but
// may be used for other Layout types as well.  Check the
// documentation for the UI implementation you're using.
func LayoutName(name string) LayoutOpt {
	return func(c LayoutCfg) LayoutCfg {
		c.Name = &name
		return c
	}
}

// LayoutResize is a placeholder LayoutOpt that will later
// be used to set resize functionality for the element being
// added.
func LayoutResize() LayoutOpt {
	return func(c LayoutCfg) LayoutCfg { return c }
}

// Window represents a UI window.
type Window interface {
	HandleInput(input.Handler)
	SetChild(interface{}) error
	Size() Size
}

// Layout represents a UI element that mainly lays out other
// elements.
type Layout interface {
	Add(interface{}, ...LayoutOpt) error
	Remove(interface{}) error
	SetFocus(interface{}) error
}

// TextDisplay represents a UI element which displays text to the
// user.
type TextDisplay interface {
	// Apply applies edits to the text of the TextContainer.
	Apply(...input.Edit)

	// SetText changes the text of the TextContainer.
	SetText([]rune)

	// Text returns the current text of the TextContainer.
	Text() []rune

	// SetSpans applies a slice of spans to the text in order to
	// color it.  In cases of overlapping spans, the last span will
	// overwrite previous spans.
	//
	// If an error occurs, the element will attempt to apply as
	// many spans as it can before returning the error.
	SetSpans(...Span) error

	// Spans returns the current spans applied to the text.
	Spans() []Span
}

// TextEditor represents a UI element with user-changeable text.
type TextEditor interface {
	TextDisplay

	// SetCarets applies a caret to each index passed in, indicating
	// to the user that the text can be edited.  It is up to the
	// controlling context to handle user input and apply edits, but
	// the UI should take care of displaying the carets.
	SetCarets(...int) error

	// Carets returns the current list of caret indexes.
	Carets() []int
}
