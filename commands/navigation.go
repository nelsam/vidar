// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

type PrevLine struct {
}

func NewPrevLine() *PrevLine {
	return &PrevLine{}
}

func (PrevLine) Name() string {
	return "prev-line"
}

func (PrevLine) Menu() string {
	return "Navigation"
}

func (PrevLine) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	finder.CurrentEditor().Controller().MoveUp()
	return true, true
}

type NextLine struct {
}

func NewNextLine() *NextLine {
	return &NextLine{}
}

func (NextLine) Name() string {
	return "next-line"
}

func (NextLine) Menu() string {
	return "Navigation"
}

func (NextLine) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	finder.CurrentEditor().Controller().MoveDown()
	return true, true
}

type PrevChar struct {
}

func NewPrevChar() *PrevChar {
	return &PrevChar{}
}

func (PrevChar) Name() string {
	return "prev-char"
}

func (PrevChar) Menu() string {
	return "Navigation"
}

func (PrevChar) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	finder.CurrentEditor().Controller().MoveLeft()
	return true, true
}

type NextChar struct {
}

func NewNextChar() *NextChar {
	return &NextChar{}
}

func (NextChar) Name() string {
	return "next-char"
}

func (NextChar) Menu() string {
	return "Navigation"
}

func (NextChar) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	finder.CurrentEditor().Controller().MoveRight()
	return true, true
}

type EndOfLine struct {
}

func NewEndOfLine() *EndOfLine {
	return &EndOfLine{}
}

func (EndOfLine) Name() string {
	return "end-of-line"
}

func (EndOfLine) Menu() string {
	return "Navigation"
}

func (EndOfLine) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	finder.CurrentEditor().Controller().MoveEnd()
	return true, true
}

type BeginningOfLine struct {
}

func NewBeginningOfLine() *BeginningOfLine {
	return &BeginningOfLine{}
}

func (BeginningOfLine) Name() string {
	return "beginning-of-line"
}

func (BeginningOfLine) Menu() string {
	return "Navigation"
}

func (BeginningOfLine) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	finder.CurrentEditor().Controller().MoveHome()
	return true, true
}
