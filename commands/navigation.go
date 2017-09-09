// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commands/cursor"
)

type Mover interface {
	For(cursor.Direction, cursor.Mod) *cursor.Mover
}

type Executor interface {
	Execute(bind.Bindable)
}

type Scroller interface {
	ScrollToRune(int)
}

type Careter interface {
	FirstCaret() int
	Deselect(bool) bool
}

type Scroll struct {
	name      string
	direction cursor.Direction
	mod       cursor.Mod

	mover    Mover
	exec     Executor
	scroller Scroller
	careter  Careter
}

func (s *Scroll) Name() string {
	return s.name
}

func (s *Scroll) Menu() string {
	return "Navigation"
}

func (s *Scroll) Reset() {
	s.mover = nil
	s.exec = nil
	s.scroller = nil
	s.careter = nil
}

func (s *Scroll) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case Mover:
		s.mover = src
	case Executor:
		s.exec = src
	case Scroller:
		s.scroller = src
	case Careter:
		s.careter = src
	}
	if s.mover != nil && s.exec != nil && s.scroller != nil && s.careter != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (s *Scroll) Exec() error {
	m := s.mover.For(s.direction, s.mod)
	s.exec.Execute(m)
	s.scroller.ScrollToRune(s.careter.FirstCaret())
	return nil
}

type ScrollDeselect struct {
	Scroll
}

func (s *ScrollDeselect) Exec() error {
	if err := s.Scroll.Exec(); err != nil {
		return err
	}
	s.careter.Deselect(false)
	return nil
}

type NavHook struct {
}

func (n NavHook) Name() string {
	return "navigation-hook"
}

func (n NavHook) OpName() string {
	return "open-file"
}

func (n NavHook) FileBindables(string) []bind.Bindable {
	return []bind.Bindable{
		&cursor.Mover{},
		NewPrevLine(),
		NewSelectPrevLine(),
		NewNextLine(),
		NewSelectNextLine(),
		NewPrevChar(),
		NewPrevWord(),
		NewSelectPrevChar(),
		NewSelectPrevWord(),
		NewNextChar(),
		NewNextWord(),
		NewSelectNextChar(),
		NewSelectNextWord(),
		NewLineStart(),
		NewSelectLineStart(),
		NewLineEnd(),
		NewSelectLineEnd(),
	}
}

func NewPrevLine() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "prev-line",
			direction: cursor.Up,
		},
	}
}

func NewSelectPrevLine() bind.Command {
	return &Scroll{
		name:      "select-prev-line",
		direction: cursor.Up,
		mod:       cursor.Select,
	}
}

func NewNextLine() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "next-line",
			direction: cursor.Down,
		},
	}
}

func NewSelectNextLine() bind.Command {
	return &Scroll{
		name:      "select-next-line",
		direction: cursor.Down,
		mod:       cursor.Select,
	}
}

func NewPrevChar() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "prev-char",
			direction: cursor.Left,
		},
	}
}

func NewPrevWord() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "prev-word",
			direction: cursor.Left,
			mod:       cursor.Word,
		},
	}
}

func NewSelectPrevChar() bind.Command {
	return &Scroll{
		name:      "select-prev-char",
		direction: cursor.Left,
		mod:       cursor.Select,
	}
}

func NewSelectPrevWord() bind.Command {
	return &Scroll{
		name:      "select-prev-word",
		direction: cursor.Left,
		mod:       cursor.Select | cursor.Word,
	}
}

func NewNextChar() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "next-char",
			direction: cursor.Right,
		},
	}
}

func NewNextWord() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "next-word",
			direction: cursor.Right,
			mod:       cursor.Word,
		},
	}
}

func NewSelectNextChar() bind.Command {
	return &Scroll{
		name:      "select-next-char",
		direction: cursor.Right,
		mod:       cursor.Select,
	}
}

func NewSelectNextWord() bind.Command {
	return &Scroll{
		name:      "select-next-word",
		direction: cursor.Right,
		mod:       cursor.Select | cursor.Word,
	}
}

func NewLineEnd() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "line-end",
			direction: cursor.Right,
			mod:       cursor.Line,
		},
	}
}

func NewSelectLineEnd() bind.Command {
	return &Scroll{
		name:      "select-to-line-end",
		direction: cursor.Right,
		mod:       cursor.SelectLine,
	}
}

func NewLineStart() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "line-start",
			direction: cursor.Left,
			mod:       cursor.Line,
		},
	}
}

func NewSelectLineStart() bind.Command {
	return &Scroll{
		name:      "select-to-line-start",
		direction: cursor.Left,
		mod:       cursor.SelectLine,
	}
}
