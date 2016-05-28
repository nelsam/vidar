// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import "github.com/nelsam/vidar/editor"

type EditorCommand interface {
	Name() string
	Menu() string
	Exec(*editor.CodeEditor)
}

type EditorExecutor struct {
	EditorCommand
}

func (e EditorExecutor) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	e.EditorCommand.Exec(finder.CurrentEditor())
	return true, true
}

type Scroller struct {
	EditorCommand
}

func NewScroller(cmd EditorCommand) EditorExecutor {
	return EditorExecutor{
		EditorCommand: Scroller{
			EditorCommand: cmd,
		},
	}
}

func (s Scroller) Exec(editor *editor.CodeEditor) {
	s.EditorCommand.Exec(editor)
	editor.ScrollToRune(editor.Controller().FirstCaret())
}

type Mover struct {
	EditorCommand
}

func NewMover(cmd EditorCommand) EditorExecutor {
	return NewScroller(Mover{EditorCommand: cmd})
}

func (m Mover) Exec(editor *editor.CodeEditor) {
	if editor.Controller().Deselect(true) {
		return
	}
	m.EditorCommand.Exec(editor)
}

type PrevLine struct {
}

func NewPrevLine() EditorExecutor {
	return NewMover(PrevLine{})
}

func (PrevLine) Name() string {
	return "prev-line"
}

func (PrevLine) Menu() string {
	return "Navigation"
}

func (PrevLine) Exec(editor *editor.CodeEditor) {
	editor.Controller().MoveUp()
}

type SelectPrevLine struct {
}

func NewSelectPrevLine() EditorExecutor {
	return NewScroller(SelectPrevLine{})
}

func (SelectPrevLine) Name() string {
	return "select-prev-line"
}

func (SelectPrevLine) Menu() string {
	return "Navigation"
}

func (SelectPrevLine) Exec(editor *editor.CodeEditor) {
	editor.Controller().SelectUp()
}

type NextLine struct {
}

func NewNextLine() EditorExecutor {
	return NewMover(NextLine{})
}

func (NextLine) Name() string {
	return "next-line"
}

func (NextLine) Menu() string {
	return "Navigation"
}

func (NextLine) Exec(editor *editor.CodeEditor) {
	editor.Controller().MoveDown()
}

type SelectNextLine struct {
}

func NewSelectNextLine() EditorExecutor {
	return NewScroller(SelectNextLine{})
}

func (SelectNextLine) Name() string {
	return "select-next-line"
}

func (SelectNextLine) Menu() string {
	return "Navigation"
}

func (SelectNextLine) Exec(editor *editor.CodeEditor) {
	editor.Controller().SelectDown()
}

type PrevChar struct {
}

func NewPrevChar() EditorExecutor {
	return NewMover(PrevChar{})
}

func (PrevChar) Name() string {
	return "prev-char"
}

func (PrevChar) Menu() string {
	return "Navigation"
}

func (PrevChar) Exec(editor *editor.CodeEditor) {
	editor.Controller().MoveLeft()
}

type SelectPrevChar struct {
}

func NewSelectPrevChar() EditorExecutor {
	return NewScroller(SelectPrevChar{})
}

func (SelectPrevChar) Name() string {
	return "select-prev-char"
}

func (SelectPrevChar) Menu() string {
	return "Navigation"
}

func (SelectPrevChar) Exec(editor *editor.CodeEditor) {
	editor.Controller().SelectLeft()
}

type NextChar struct {
}

func NewNextChar() EditorExecutor {
	return NewMover(NextChar{})
}

func (NextChar) Name() string {
	return "next-char"
}

func (NextChar) Menu() string {
	return "Navigation"
}

func (NextChar) Exec(editor *editor.CodeEditor) {
	editor.Controller().MoveRight()
}

type SelectNextChar struct {
}

func NewSelectNextChar() EditorExecutor {
	return NewScroller(SelectNextChar{})
}

func (SelectNextChar) Name() string {
	return "select-next-char"
}

func (SelectNextChar) Menu() string {
	return "Navigation"
}

func (SelectNextChar) Exec(editor *editor.CodeEditor) {
	editor.Controller().SelectRight()
}

type LineEnd struct {
}

func NewLineEnd() EditorExecutor {
	return NewMover(LineEnd{})
}

func (LineEnd) Name() string {
	return "line-end"
}

func (LineEnd) Menu() string {
	return "Navigation"
}

func (LineEnd) Exec(editor *editor.CodeEditor) {
	editor.Controller().MoveEnd()
}

type SelectLineEnd struct {
}

func NewSelectLineEnd() EditorExecutor {
	return NewScroller(SelectLineEnd{})
}

func (SelectLineEnd) Name() string {
	return "select-to-line-end"
}

func (SelectLineEnd) Menu() string {
	return "Navigation"
}

func (SelectLineEnd) Exec(editor *editor.CodeEditor) {
	editor.Controller().SelectEnd()
}

type LineStart struct {
}

func NewLineStart() EditorExecutor {
	return NewMover(LineStart{})
}

func (LineStart) Name() string {
	return "line-start"
}

func (LineStart) Menu() string {
	return "Navigation"
}

func (LineStart) Exec(editor *editor.CodeEditor) {
	editor.Controller().MoveHome()
}

type SelectLineStart struct {
}

func NewSelectLineStart() EditorExecutor {
	return NewScroller(SelectLineStart{})
}

func (SelectLineStart) Name() string {
	return "select-to-line-start"
}

func (SelectLineStart) Menu() string {
	return "Navigation"
}

func (SelectLineStart) Exec(editor *editor.CodeEditor) {
	editor.Controller().SelectHome()
}
