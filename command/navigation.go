// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/command/caret"
	"github.com/nelsam/vidar/plugin/command"
)

type Mover interface {
	For(caret.Direction, caret.Mod) *caret.Mover
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
	direction caret.Direction
	mod       caret.Mod

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
	Commander command.Commander
}

func (n NavHook) Name() string {
	return "navigation-hook"
}

func (n NavHook) OpName() string {
	return "open-file"
}

func (n NavHook) FileBindables(string) []bind.Bindable {
	return []bind.Bindable{
		&caret.Mover{},
		&caret.OnEdit{Commander: n.Commander},
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
			direction: caret.Up,
		},
	}
}

func NewSelectPrevLine() bind.Command {
	return &Scroll{
		name:      "select-prev-line",
		direction: caret.Up,
		mod:       caret.Select,
	}
}

func NewNextLine() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "next-line",
			direction: caret.Down,
		},
	}
}

func NewSelectNextLine() bind.Command {
	return &Scroll{
		name:      "select-next-line",
		direction: caret.Down,
		mod:       caret.Select,
	}
}

func NewPrevChar() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "prev-char",
			direction: caret.Left,
		},
	}
}

func NewPrevWord() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "prev-word",
			direction: caret.Left,
			mod:       caret.Word,
		},
	}
}

func NewSelectPrevChar() bind.Command {
	return &Scroll{
		name:      "select-prev-char",
		direction: caret.Left,
		mod:       caret.Select,
	}
}

func NewSelectPrevWord() bind.Command {
	return &Scroll{
		name:      "select-prev-word",
		direction: caret.Left,
		mod:       caret.Select | caret.Word,
	}
}

func NewNextChar() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "next-char",
			direction: caret.Right,
		},
	}
}

func NewNextWord() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "next-word",
			direction: caret.Right,
			mod:       caret.Word,
		},
	}
}

func NewSelectNextChar() bind.Command {
	return &Scroll{
		name:      "select-next-char",
		direction: caret.Right,
		mod:       caret.Select,
	}
}

func NewSelectNextWord() bind.Command {
	return &Scroll{
		name:      "select-next-word",
		direction: caret.Right,
		mod:       caret.Select | caret.Word,
	}
}

func NewLineEnd() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "line-end",
			direction: caret.Right,
			mod:       caret.Line,
		},
	}
}

func NewSelectLineEnd() bind.Command {
	return &Scroll{
		name:      "select-to-line-end",
		direction: caret.Right,
		mod:       caret.SelectLine,
	}
}

func NewLineStart() bind.Command {
	return &ScrollDeselect{
		Scroll: Scroll{
			name:      "line-start",
			direction: caret.Left,
			mod:       caret.Line,
		},
	}
}

func NewSelectLineStart() bind.Command {
	return &Scroll{
		name:      "select-to-line-start",
		direction: caret.Left,
		mod:       caret.SelectLine,
	}
}
