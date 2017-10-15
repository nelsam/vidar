// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"fmt"
	"go/token"
	"log"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/mixins/outer"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/theme"
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

type Opener interface {
	bind.Bindable
	SetLocation(string, token.Position)
}

type Commander interface {
	Bindable(string) bind.Bindable
	Execute(bind.Bindable)
}

type MultiEditor interface {
	gxui.Control
	outer.LayoutChildren
	Has(hiddenPrefix, path string) bool
	Open(hiddenPrefix, path, headerText string, environ []string) (editor *CodeEditor, existed bool)
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

	driver      gxui.Driver
	cmdr        Commander
	theme       *basic.Theme
	syntaxTheme theme.Theme
	font        gxui.Font
	window      gxui.Window

	current MultiEditor
}

func NewSplitEditor(driver gxui.Driver, cmdr Commander, window gxui.Window, theme *basic.Theme, syntaxTheme theme.Theme, font gxui.Font) *SplitEditor {
	editor := &SplitEditor{
		driver:      driver,
		cmdr:        cmdr,
		theme:       theme,
		syntaxTheme: syntaxTheme,
		font:        font,
		window:      window,
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
	newSplit := NewTabbedEditor(e.driver, e.cmdr, e.theme, e.syntaxTheme, e.font)
	defer func() {
		newSplit.Add(name, editor)
		opener := e.cmdr.Bindable("open-file").(Opener)
		opener.SetLocation(editor.Filepath(), token.Position{})
		e.cmdr.Execute(opener)
	}()
	if e.Orientation() == orientation {
		e.AddChild(newSplit)
		return
	}
	newSplitter := NewSplitEditor(e.driver, e.cmdr, e.window, e.theme, e.syntaxTheme, e.font)
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
		opener := e.cmdr.Bindable("open-file").(Opener)
		opener.SetLocation(e.current.CurrentEditor().Filepath(), token.Position{})
		e.cmdr.Execute(opener)
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
		opener := e.cmdr.Bindable("open-file").(Opener)
		opener.SetLocation(newFocus.CurrentEditor().Filepath(), token.Position{})
		e.cmdr.Execute(opener)
		break
	}
	e.SplitterLayout.MouseUp(event)
}

func (e *SplitEditor) Has(hiddenPrefix, path string) bool {
	for _, child := range e.Children() {
		if me, ok := child.Control.(MultiEditor); ok && me.Has(hiddenPrefix, path) {
			return true
		}
	}
	return false
}

func (e *SplitEditor) Open(hiddenPrefix, path, headerText string, environ []string) (editor *CodeEditor, existed bool) {
	for _, child := range e.Children() {
		if me, ok := child.Control.(MultiEditor); ok && me.Has(hiddenPrefix, path) {
			e.current = me
			return me.Open(hiddenPrefix, path, headerText, environ)
		}
	}
	return e.current.Open(hiddenPrefix, path, headerText, environ)
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

func (e *SplitEditor) NextEditor(direction Direction) *CodeEditor {
	editor, _ := e.nextEditor(direction)
	return editor
}

func (e *SplitEditor) nextEditor(direction Direction) (editor *CodeEditor, wrapped bool) {
	switch direction {
	case Up, Down:
		if e.Orientation().Horizontal() {
			if splitter, ok := e.current.(*SplitEditor); ok {
				return splitter.nextEditor(direction)
			}
			return nil, false
		}
	case Left, Right:
		if e.Orientation().Vertical() {
			if splitter, ok := e.current.(*SplitEditor); ok {
				return splitter.nextEditor(direction)
			}
			return nil, false
		}
	}

	if splitter, ok := e.current.(*SplitEditor); ok {
		// Special case - there could be another split editor with our orientation
		// as a child, and *that* is the editor that needs the split moved.
		ed, wrapped := splitter.nextEditor(direction)
		if e != nil && !wrapped {
			return ed, false
		}
	}

	children := e.Children()
	i := children.IndexOf(e.current)
	if i < 0 {
		log.Printf("Error: Current editor is not part of the splitter's layout")
		return nil, false
	}
	var next func(i int) int
	switch direction {
	case Up, Left:
		next = func(i int) int {
			i--
			if i < 0 {
				wrapped = true
				i = len(children) - 1
			}
			return i
		}
	case Down, Right:
		next = func(i int) int {
			i++
			if i == len(children) {
				wrapped = true
				i = 0
			}
			return i
		}
	}
	var (
		me MultiEditor
		ok bool
	)

	for {
		i = next(i)
		me, ok = children[i].Control.(MultiEditor)
		if ok {
			break
		}
	}
	if splitter, ok := me.(*SplitEditor); ok {
		return splitter.first(direction), wrapped
	}

	return me.CurrentEditor(), wrapped
}

func (e *SplitEditor) first(d Direction) *CodeEditor {
	switch d {
	case Up, Down:
		if e.Orientation().Horizontal() {
			if splitter, ok := e.current.(*SplitEditor); ok {
				return splitter.first(d)
			}
			return e.current.CurrentEditor()
		}
	case Left, Right:
		if e.Orientation().Vertical() {
			if splitter, ok := e.current.(*SplitEditor); ok {
				return splitter.first(d)
			}
			return e.current.CurrentEditor()
		}
	}

	var first *gxui.Child
	switch d {
	case Up, Left:
		first = e.Children()[0]
	case Down, Right:
		first = e.Children()[len(e.Children())-1]
	}

	switch src := first.Control.(type) {
	case *SplitEditor:
		return src.first(d)
	case MultiEditor:
		return src.CurrentEditor()
	default:
		log.Printf("Error: first editor is not an editor")
		return nil
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

type SplitterBar struct {
	mixins.SplitterBar
	viewport    gxui.Viewport
	orientation gxui.Orientation

	arrow, horizResize, vertResize *glfw.Cursor
}

func NewSplitterBar(viewport gxui.Viewport, theme gxui.Theme) *SplitterBar {
	s := &SplitterBar{
		viewport:    viewport,
		arrow:       glfw.CreateStandardCursor(glfw.ArrowCursor),
		horizResize: glfw.CreateStandardCursor(glfw.HResizeCursor),
		vertResize:  glfw.CreateStandardCursor(glfw.VResizeCursor),
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
