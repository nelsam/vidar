// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"fmt"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/mixins/outer"
	"github.com/nelsam/gxui/themes/basic"
)

var (
	splitterBG = gxui.Color{
		R: .05,
		G: .05,
		B: .05,
		A: 1,
	}
	splitterFG = gxui.Color{
		R: .2,
		G: .2,
		B: .2,
		A: 1,
	}
)

type Orienter interface {
	SetOrientation(gxui.Orientation)
}

type Splitter interface {
	Split(orientation gxui.Orientation)
}

type MultiEditor interface {
	gxui.Control
	outer.LayoutChildren
	Focus()
	Open(name, path, headerText string, environ []string) *CodeEditor
	Editors() uint
	CurrentEditor() *CodeEditor
	CurrentFile() string
	CloseCurrentEditor() (name string, editor *CodeEditor)
	Add(name string, editor *CodeEditor)
	SaveAll()
}

type Direction int

const (
	Up Direction = 1 + iota
	Right
	Down
	Left
)

type SplitEditor struct {
	mixins.SplitterLayout

	driver gxui.Driver
	theme  *basic.Theme
	font   gxui.Font
	window gxui.Window

	current MultiEditor
}

func NewSplitEditor(driver gxui.Driver, window gxui.Window, theme *basic.Theme, font gxui.Font) *SplitEditor {
	editor := &SplitEditor{
		driver: driver,
		theme:  theme,
		font:   font,
		window: window,
	}
	editor.SplitterLayout.Init(editor, theme)
	return editor
}

func (e *SplitEditor) Elements() []interface{} {
	return []interface{}{e.current}
}

func (e *SplitEditor) Split(orientation gxui.Orientation) {
	if e.current.Editors() <= 1 {
		return
	}
	if splitter, ok := e.current.(Splitter); ok {
		splitter.Split(orientation)
		return
	}
	name, editor := e.current.CloseCurrentEditor()
	newSplit := NewTabbedEditor(e.driver, e.theme, e.font)
	defer func() {
		newSplit.Add(name, editor)
		newSplit.Focus()
	}()
	if e.Orientation() == orientation {
		e.AddChild(newSplit)
		return
	}
	newSplitter := NewSplitEditor(e.driver, e.window, e.theme, e.font)
	newSplitter.SetOrientation(orientation)
	var (
		index       int
		searchChild *gxui.Child
	)
	for index, searchChild = range e.Children() {
		if e.current == searchChild.Control {
			break
		}
	}
	e.RemoveChildAt(index)
	newSplitter.AddChild(e.current)
	e.current = newSplitter
	e.AddChildAt(index, e.current)
	newSplitter.AddChild(newSplit)
}

func (e *SplitEditor) Editors() (count uint) {
	for _, child := range e.Children() {
		editor, ok := child.Control.(MultiEditor)
		if !ok {
			continue
		}
		count += editor.Editors()
	}
	return count
}

func (e *SplitEditor) CloseCurrentEditor() (name string, editor *CodeEditor) {
	name, editor = e.current.CloseCurrentEditor()
	if e.current.Editors() == 0 && len(e.Children()) > 1 {
		e.RemoveChild(e.current)
		e.current = e.Children()[0].Control.(MultiEditor)
		e.current.Focus()
	}
	return name, editor
}

func (e *SplitEditor) Add(name string, editor *CodeEditor) {
	e.current.Add(name, editor)
}

func (e *SplitEditor) AddChild(child gxui.Control) *gxui.Child {
	editor, ok := child.(MultiEditor)
	if !ok {
		panic(fmt.Errorf("SplitEditor: Non-MultiEditor type %T sent to AddChild", child))
	}
	if e.current == nil {
		e.current = editor
	}
	return e.SplitterLayout.AddChild(child)
}

func (e *SplitEditor) Focus() {
	e.current.Focus()
}

func (l *SplitEditor) CreateSplitterBar() gxui.Control {
	b := NewSplitterBar(l.window.Viewport(), l.theme)
	b.SetOrientation(l.Orientation())
	b.OnSplitterDragged(func(wndPnt math.Point) { l.SplitterDragged(b, wndPnt) })
	return b
}

func (e *SplitEditor) SetOrientation(o gxui.Orientation) {
	e.SplitterLayout.SetOrientation(o)
	for _, child := range e.Children() {
		if orienter, ok := child.Control.(Orienter); ok {
			orienter.SetOrientation(o)
		}
	}
}

func (e *SplitEditor) MouseUp(event gxui.MouseEvent) {
	for _, child := range e.Children() {
		offsetPoint := event.Point.AddX(-child.Offset.X).AddY(-child.Offset.Y)
		if !child.Control.ContainsPoint(offsetPoint) {
			continue
		}
		newFocus, ok := child.Control.(MultiEditor)
		if !ok {
			continue
		}
		e.current = newFocus
		e.current.Focus()
		break
	}
	e.SplitterLayout.MouseUp(event)
}

func (e *SplitEditor) Open(name, path, headerText string, environ []string) *CodeEditor {
	return e.current.Open(name, path, headerText, environ)
}

func (e *SplitEditor) CurrentEditor() *CodeEditor {
	return e.current.CurrentEditor()
}

func (e *SplitEditor) CurrentFile() string {
	return e.current.CurrentFile()
}

func (e *SplitEditor) ChildIndex(c gxui.Control) int {
	if c == nil {
		return -1
	}
	for i, child := range e.Children() {
		if child.Control == c {
			return i
		}
	}
	return -1
}

func (e *SplitEditor) ShiftSplit(direction Direction) {
	switch direction {
	case Up, Down:
		if e.Orientation().Horizontal() {
			if splitter, ok := e.current.(*SplitEditor); ok {
				splitter.ShiftSplit(direction)
			}
			return
		}
	case Left, Right:
		if e.Orientation().Vertical() {
			if splitter, ok := e.current.(*SplitEditor); ok {
				splitter.ShiftSplit(direction)
			}
			return
		}
	}

	switch direction {
	case Up:
		if !e.shiftUp() {
			e.focus(Down)
		}
	case Down:
		if !e.shiftDown() {
			e.focus(Up)
		}
	case Left:
		if !e.shiftLeft() {
			e.focus(Right)
		}
	case Right:
		if !e.shiftRight() {
			e.focus(Left)
		}
	}
}

func (e *SplitEditor) SaveAll() {
	for _, child := range e.Children() {
		editor, ok := child.Control.(MultiEditor)
		if !ok {
			continue
		}
		editor.SaveAll()
	}
}

func (e *SplitEditor) focusFirst() {
	e.current = e.Children()[0].Control.(MultiEditor)
}

func (e *SplitEditor) focusLast() {
	e.current = e.Children()[len(e.Children())-1].Control.(MultiEditor)
}

func (e *SplitEditor) focus(d Direction) {
	defer func() {
		if splitter, ok := e.current.(*SplitEditor); ok {
			splitter.focus(d)
		}
	}()

	switch d {
	case Left, Right:
		if e.Orientation().Vertical() {
			return
		}
	case Up, Down:
		if e.Orientation().Horizontal() {
			return
		}
	}
	switch d {
	case Up, Left:
		e.focusFirst()
	case Down, Right:
		e.focusLast()
	}
}

func (e *SplitEditor) focusIndex(idx int, d Direction) bool {
	editor, ok := e.Children()[idx].Control.(MultiEditor)
	if !ok {
		return false
	}
	if splitter, ok := editor.(*SplitEditor); ok {
		splitter.focus(d)
	}
	e.current = editor
	return true
}

func (e *SplitEditor) shiftUp() bool {
	if e.Orientation().Horizontal() {
		if editor, ok := e.current.(*SplitEditor); ok {
			return editor.shiftUp()
		}
		return false
	}

	if splitter, ok := e.current.(*SplitEditor); ok {
		if splitter.shiftUp() {
			return true
		}
	}
	for idx := e.ChildIndex(e.current) - 1; idx >= 0; idx-- {
		if e.focusIndex(idx, Down) {
			return true
		}
	}
	return false
}

func (e *SplitEditor) shiftDown() bool {
	if e.Orientation().Horizontal() {
		if editor, ok := e.current.(*SplitEditor); ok {
			return editor.shiftDown()
		}
		return false
	}

	if splitter, ok := e.current.(*SplitEditor); ok {
		if splitter.shiftDown() {
			return true
		}
	}
	for idx := e.ChildIndex(e.current) + 1; idx < len(e.Children()); idx++ {
		if e.focusIndex(idx, Up) {
			return true
		}
	}
	return false
}

func (e *SplitEditor) shiftLeft() bool {
	if e.Orientation().Vertical() {
		if editor, ok := e.current.(*SplitEditor); ok {
			return editor.shiftLeft()
		}
		return false
	}

	if splitter, ok := e.current.(*SplitEditor); ok {
		if splitter.shiftLeft() {
			return true
		}
	}
	for idx := e.ChildIndex(e.current) - 1; idx >= 0; idx-- {
		if e.focusIndex(idx, Right) {
			return true
		}
	}
	return false
}

func (e *SplitEditor) shiftRight() bool {
	if e.Orientation().Vertical() {
		if editor, ok := e.current.(*SplitEditor); ok {
			return editor.shiftRight()
		}
		return false
	}

	if splitter, ok := e.current.(*SplitEditor); ok {
		if splitter.shiftRight() {
			return true
		}
	}
	for idx := e.ChildIndex(e.current) + 1; idx < len(e.Children()); idx++ {
		if e.focusIndex(idx, Left) {
			return true
		}
	}
	return false
}

type SplitterBar struct {
	mixins.SplitterBar
	viewport    gxui.Viewport
	orientation gxui.Orientation

	arrow, horizResize, vertResize *glfw.Cursor
}

func NewSplitterBar(viewport gxui.Viewport, theme gxui.Theme) *SplitterBar {
	s := &SplitterBar{
		viewport:    viewport,
		arrow:       glfw.CreateStandardCursor(int(glfw.ArrowCursor)),
		horizResize: glfw.CreateStandardCursor(int(glfw.HResizeCursor)),
		vertResize:  glfw.CreateStandardCursor(int(glfw.VResizeCursor)),
	}
	s.SplitterBar.Init(s, theme)
	s.SetBackgroundColor(splitterBG)
	s.SetForegroundColor(splitterFG)
	return s
}

func (s *SplitterBar) SetOrientation(o gxui.Orientation) {
	s.orientation = o
}

func (s *SplitterBar) MouseEnter(gxui.MouseEvent) {
	switch s.orientation {
	case gxui.Vertical:
		s.viewport.SetCursor(s.vertResize)
	case gxui.Horizontal:
		s.viewport.SetCursor(s.horizResize)
	}
}

func (s *SplitterBar) MouseExit(gxui.MouseEvent) {
	s.viewport.SetCursor(s.arrow)
}
