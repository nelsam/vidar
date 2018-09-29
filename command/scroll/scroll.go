// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package scroll

import (
	"errors"
	"fmt"

	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
)

// Controller matches the type we need to control scrolling.
type Controller interface {
	ScrollToLine(int)
	ScrollToRune(int)
}

// ScrolledHook is a hook that will trigger whenever the view has
// scrolled.
type ScrolledHook interface {
	Scrolled(e input.Editor, focus int)
}

// Opt is an option function to be passed to Scroller.For
type Opt func(*Scroller) *Scroller

// clone just creates a clone of s.
func clone(s *Scroller) *Scroller {
	clone := &Scroller{
		pos:      s.pos,
		line:     s.line,
		scrolled: s.scrolled,
		editor:   s.editor,
		ctrl:     s.ctrl,
	}
	for _, h := range s.scrolled {
		clone.scrolled = append(clone.scrolled, h)
	}
	return clone
}

// ToLine is an Opt that sets s to scroll vertically, focusing on line.
// The horizontal position will not change.
func ToLine(line int) Opt {
	return func(s *Scroller) *Scroller {
		newS := clone(s)
		newS.line = line
		newS.pos = -1
		return newS
	}
}

// ToRune is an Opt that sets s to scroll directly to a rune position.
func ToRune(pos int) Opt {
	return func(s *Scroller) *Scroller {
		newS := clone(s)
		newS.line = -1
		newS.pos = pos
		return newS
	}
}

// Scroller is a bind.Op that is used to scroll the editor frame.
type Scroller struct {
	pos      int
	line     int
	scrolled []ScrolledHook

	editor input.Editor
	ctrl   Controller
}

// Name returns Scroller's name.
func (*Scroller) Name() string {
	return "scroll"
}

// For returns a copy of s (as a bind.Bindable for dependency reasons) which
// has had the passed in opts applied.
func (s *Scroller) For(opts ...Opt) bind.Bindable {
	for _, o := range opts {
		s = o(s)
	}
	return s
}

// Bind clones s, binds b as a hook to the clone, and returns it.
func (s *Scroller) Bind(b bind.Bindable) (bind.HookedMultiOp, error) {
	newS := clone(s)

	if h, ok := b.(ScrolledHook); ok {
		newS.scrolled = append(newS.scrolled, h)
		return newS, nil
	}
	return nil, fmt.Errorf("hook %s does not implement ScrolledHook", b.Name())
}

// Reset resets the execution state of s.
func (s *Scroller) Reset() {
	s.editor = nil
	s.ctrl = nil
}

// Store checks elem for the necessary methods and stores it if it's needed.
func (s *Scroller) Store(elem interface{}) bind.Status {
	// We can't do a type switch here because with gxui, at least,
	// the Controller is also the input.Editor.
	if e, ok := elem.(input.Editor); ok {
		s.editor = e
	}
	if c, ok := elem.(Controller); ok {
		s.ctrl = c
	}

	if s.editor != nil && s.ctrl != nil {
		return bind.Done
	}
	return bind.Waiting
}

// Exec executes s after it has stored all the values it needs.
func (s *Scroller) Exec() error {
	switch {
	case s.line >= 0:
		s.ctrl.ScrollToLine(s.line)
	case s.pos >= 0:
		s.ctrl.ScrollToRune(s.pos)
	default:
		return errors.New("scroll executed without any position information set")
	}
	return nil
}
