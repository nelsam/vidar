// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package cursor

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
	SetCarets([]int)
}

type MovingHook interface {
	Moving(input.Editor, Direction, []int) Direction
}

type MovedHook interface {
	Moved(input.Editor, Direction, []int)
}

type Direction int

const (
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
	moving    []MovingHook
	moved     []MovedHook

	editor input.Editor
	ctrl   Controller
}

func (*Mover) Name() string {
	return "cursor-movement"
}

func (*Mover) OpName() string {
	return "input-handler"
}

func (m *Mover) Applied(e input.Editor, edits []input.Edit) {
	h := e.(CaretHandler)
	carets := h.Carets()
	for _, e := range edits {
		carets = m.moveCarets(carets, e)
	}
	h.SetCarets(carets)
}

func (m *Mover) moveCarets(carets []int, e input.Edit) []int {
	delta := len(e.New) - len(e.Old)
	if delta == 0 {
		return carets
	}
	for i, c := range carets {
		if c >= e.At {
			carets[i] = c + delta
		}
	}
	return carets
}

func (m *Mover) For(dir Direction, mod Mod) *Mover {
	return &Mover{
		direction: dir,
		mod:       mod,
		moving:    m.moving,
		moved:     m.moved,
		editor:    m.editor,
		ctrl:      m.ctrl,
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

func (m *Mover) Exec() error {
	for _, h := range m.moving {
		m = m.For(h.Moving(m.editor, m.direction, m.ctrl.Carets()), m.mod)
	}
	switch m.direction {
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
		h.Moved(m.editor, m.direction, m.ctrl.Carets())
	}
	return nil
}
