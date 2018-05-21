// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package scroll

import (
	"fmt"

	"github.com/nelsam/gxui/math"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/editor"
)

type Controller interface {
	ScrollToLine(int)
	SetScrollOffset(int)
}

type ScrolledHook interface {
	Scrolled(e input.Editor, d Direction, pos int)
}

type Direction int

const (
	ToLine Direction = iota
	ToRune
	SetOffset
)

type Scroller struct {
	direction Direction
	position  int
	scrolled  ScrolledHook

	editor input.Editor
	ctrl   Controller
}

func (*Scroller) Name() string {
	return "scroll-setting"
}

func (m *Scroller) To(dir Direction, pos int) bind.Bindable {
	return &Scroller{
		position:  pos,
		direction: dir,
		scrolled:  m.scrolled,
		editor:    m.editor,
		ctrl:      m.ctrl,
	}
}

func (m *Scroller) Bind(b bind.Bindable) (bind.HookedMultiOp, error) {

	newS := &Scroller{
		direction: m.direction,
	}

	if h, ok := b.(ScrolledHook); ok {
		newS.scrolled = h
		return newS, nil
	}
	return nil, fmt.Errorf("hook %s not implemented ScrolledHook", b.Name())

}

func (m *Scroller) Reset() {
	m.editor = nil
	m.ctrl = nil
}

func (m *Scroller) Store(elem interface{}) bind.Status {
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

func (m *Scroller) Exec() error {
	editor := m.editor.(*editor.CodeEditor)
	lineIndex := math.Max(editor.LineIndex(m.position)-3, 0)
	editor.ScrollToLine(lineIndex)
	return nil
}
