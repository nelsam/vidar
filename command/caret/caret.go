// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package caret

import (
	"fmt"

	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
)

type Controller interface {
	Carets() []int
	MoveUp()
	SelectUp()
	MoveDown()
	SelectDown()
	MoveLeft()
	SelectLeft()
	MoveLeftByWord()
	SelectLeftByWord()
	MoveRight()
	SelectRight()
	MoveRightByWord()
	SelectRightByWord()
	MoveEnd()
	SelectEnd()
	MoveHome()
	SelectHome()
}

type CaretHandler interface {
	Carets() []int
	SetCarets(...int)
}

// MovingHook is a hook that needs to trigger when the caret is moving.
// If d and/or carets need to be modified by the hook before carets
// actually move, the MovingHook can return a different direction
// and/or slice of carets.
type MovingHook interface {
	Moving(e input.Editor, d Direction, m Mod, carets []int) (newD Direction, newM Mod, newCarets []int)
}

// MovedHook is a hook that needs to trigger after the carets have moved.
type MovedHook interface {
	Moved(e input.Editor, carets []int)
}

type Direction int

const (
	// NoDirection means that there was no Direction used
	// to move the caret.  It does not necessarily mean
	// that it didn't move.
	NoDirection Direction = iota

	Up
	Down
	Left
	Right
)

type Mod int

const (
	Select Mod = 1 << iota
	Word
	Whitespace

	Line       = Word | Whitespace
	SelectWord = Select | Word
	SelectLine = Select | Line

	NoMod Mod = 0
)

type Mover struct {
	direction Direction
	mod       Mod
	carets    []int
	moving    []MovingHook
	moved     []MovedHook

	editor input.Editor
	ctrl   Controller
}

func (*Mover) Name() string {
	return "caret-movement"
}

func (m *Mover) For(dir Direction, mod Mod) bind.Bindable {
	return &Mover{
		direction: dir,
		mod:       mod,
		moving:    m.moving,
		moved:     m.moved,
		editor:    m.editor,
		ctrl:      m.ctrl,
	}
}

func (m *Mover) To(carets ...int) bind.Bindable {
	return &Mover{
		carets: carets,
		moving: m.moving,
		moved:  m.moved,
		editor: m.editor,
		ctrl:   m.ctrl,
	}
}

func (m *Mover) Bind(b bind.Bindable) (bind.HookedMultiOp, error) {
	newM := &Mover{
		direction: m.direction,
		mod:       m.mod,
	}
	newM.moving = append(newM.moving, m.moving...)
	newM.moved = append(newM.moved, m.moved...)

	didBind := false
	if h, ok := b.(MovingHook); ok {
		didBind = true
		newM.moving = append(newM.moving, h)
	}
	if h, ok := b.(MovedHook); ok {
		didBind = true
		newM.moved = append(newM.moved, h)
	}
	if !didBind {
		return nil, fmt.Errorf("hook %s implemented neither MovingHook nor MovedHook", b.Name())
	}
	return newM, nil
}

func (m *Mover) Reset() {
	m.editor = nil
	m.ctrl = nil
}

func (m *Mover) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case input.Editor:
		m.editor = src
	case Controller:
		m.ctrl = src
	}
	if m.editor != nil && m.ctrl != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (m *Mover) trigger(h MovingHook, d Direction, mod Mod, carets []int) bind.Bindable {
	newDir, newMod, newCarets := h.Moving(m.editor, d, mod, carets)
	if newDir == NoDirection {
		return m.To(newCarets...)
	}
	return m.For(newDir, newMod)
}

func (m *Mover) Exec() error {
	for _, h := range m.moving {
		m = m.trigger(h, m.direction, m.mod, m.carets).(*Mover)
	}
	// TODO: a lot of this is workaround BS.  This should be cleaned up soon.
	switch m.direction {
	case NoDirection:
		if m.carets == nil {
			return nil
		}
		h := m.editor.(CaretHandler)
		h.SetCarets(m.carets...)
	case Up:
		switch m.mod {
		case NoMod:
			m.ctrl.MoveUp()
		case Select:
			m.ctrl.SelectUp()
		}
	case Down:
		switch m.mod {
		case NoMod:
			m.ctrl.MoveDown()
		case Select:
			m.ctrl.SelectDown()
		}
	case Left:
		switch m.mod {
		case NoMod:
			m.ctrl.MoveLeft()
		case Select:
			m.ctrl.SelectLeft()
		case Word:
			m.ctrl.MoveLeftByWord()
		case SelectWord:
			m.ctrl.SelectLeftByWord()
		case Line:
			m.ctrl.MoveHome()
		case SelectLine:
			m.ctrl.SelectHome()
		}
	case Right:
		switch m.mod {
		case NoMod:
			m.ctrl.MoveRight()
		case Select:
			m.ctrl.SelectRight()
		case Word:
			m.ctrl.MoveRightByWord()
		case SelectWord:
			m.ctrl.SelectRightByWord()
		case Line:
			m.ctrl.MoveEnd()
		case SelectLine:
			m.ctrl.SelectEnd()
		}
	}
	for _, h := range m.moved {
		h.Moved(m.editor, m.ctrl.Carets())
	}
	return nil
}
