// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/mixins/outer"
	"github.com/nelsam/gxui/themes/basic"
)

type Splitter interface {
	Split(orientation gxui.Orientation)
}

type MultiEditor interface {
	gxui.Control
	outer.LayoutChildren
	Focus()
	Open(name, path, gopath string) *CodeEditor
	Editors() uint
	CurrentEditor() *CodeEditor
	CurrentFile() string
	CloseCurrentEditor() (name string, editor *CodeEditor)
	Add(name string, editor *CodeEditor)
}

type SplitEditor struct {
	mixins.SplitterLayout

	driver gxui.Driver
	theme  *basic.Theme
	font   gxui.Font

	current MultiEditor
}

func NewSplitEditor(driver gxui.Driver, theme *basic.Theme, font gxui.Font) *SplitEditor {
	editor := &SplitEditor{}
	editor.Init(editor, driver, theme, font)
	return editor
}

func (e *SplitEditor) Init(outer mixins.SplitterLayoutOuter, driver gxui.Driver, theme *basic.Theme, font gxui.Font) {
	e.driver = driver
	e.theme = theme
	e.font = font
	e.SplitterLayout.Init(outer, theme)
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
	if e.Orientation() == orientation {
		newSplit := NewTabbedEditor(e.driver, e.theme, e.font)
		e.AddChild(newSplit)
		newSplit.Add(name, editor)
		newSplit.Focus()
		return
	}
	newSplitter := NewSplitEditor(e.driver, e.theme, e.font)
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
	e.AddChildAt(index, newSplitter)
	newSplitter.AddChild(e.current)
	e.current = newSplitter
	e.current.Focus()
	newSplitter.Split(orientation)
}

func (e *SplitEditor) Editors() uint {
	return e.current.Editors()
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

func (e *SplitEditor) Focus() {
	e.current.Focus()
}

func (e *SplitEditor) Open(name, path, gopath string) *CodeEditor {
	if e.current == nil {
		e.current = NewTabbedEditor(e.driver, e.theme, e.font)
	}
	return e.current.Open(name, path, gopath)
}

func (e *SplitEditor) CurrentEditor() *CodeEditor {
	return e.current.CurrentEditor()
}

func (e *SplitEditor) CurrentFile() string {
	return e.current.CurrentFile()
}
