// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/suggestions"
)

type TabbedEditor struct {
	mixins.PanelHolder

	editors map[string]*CodeEditor

	driver gxui.Driver
	theme  *basic.Theme
	font   gxui.Font
}

func NewTabbedEditor(driver gxui.Driver, theme *basic.Theme, font gxui.Font) *TabbedEditor {
	editor := &TabbedEditor{}
	editor.Init(editor, driver, theme, font)
	return editor
}

func (e *TabbedEditor) Init(outer mixins.PanelHolderOuter, driver gxui.Driver, theme *basic.Theme, font gxui.Font) {
	e.editors = make(map[string]*CodeEditor)
	e.driver = driver
	e.theme = theme
	e.font = font
	e.PanelHolder.Init(outer, theme)
	e.SetMargin(math.Spacing{L: 0, T: 2, R: 0, B: 0})
}

func (e *TabbedEditor) Open(name, path, gopath, headerText string) *CodeEditor {
	if editor, ok := e.editors[name]; ok {
		e.Select(e.PanelIndex(editor))
		e.Focus()
		return editor
	}
	editor := &CodeEditor{}
	editor.Init(e.driver, e.theme, e.font, path, headerText)
	editor.SetTabWidth(4)
	suggester := suggestions.NewGoCodeProvider(editor, gopath)
	editor.SetSuggestionProvider(suggester)
	e.Add(name, editor)
	return editor
}

func (e *TabbedEditor) Add(name string, editor *CodeEditor) {
	e.editors[name] = editor
	e.AddPanel(editor, name)
	e.Select(e.PanelIndex(editor))
	e.Focus()
}

func (e *TabbedEditor) Focus() {
	if e.SelectedPanel() != nil {
		gxui.SetFocus(e.SelectedPanel().(gxui.Focusable))
	}
}

func (e *TabbedEditor) Files() []string {
	files := make([]string, 0, len(e.editors))
	for file := range e.editors {
		files = append(files, file)
	}
	return files
}

func (e *TabbedEditor) Editors() uint {
	return uint(len(e.editors))
}

func (e *TabbedEditor) CreatePanelTab() mixins.PanelTab {
	return basic.CreatePanelTab(e.theme)
}

func (e *TabbedEditor) KeyPress(event gxui.KeyboardEvent) bool {
	if event.Modifier.Control() || event.Modifier.Super() {
		switch event.Key {
		case gxui.KeyTab:
			panels := e.PanelCount()
			if panels < 2 {
				return true
			}
			current := e.PanelIndex(e.SelectedPanel())
			next := current + 1
			if event.Modifier.Shift() {
				next = current - 1
			}
			if next >= panels {
				next = 0
			}
			if next < 0 {
				next = panels - 1
			}
			e.Select(next)
			e.driver.Call(func() {
				e.Focus()
			})
			return true
		}
	}
	return e.PanelHolder.KeyPress(event)
}

func (e *TabbedEditor) CloseCurrentEditor() (name string, editor *CodeEditor) {
	toRemove := e.CurrentEditor()
	if toRemove == nil {
		return "", nil
	}
	e.RemovePanel(toRemove)
	for name, panel := range e.editors {
		if panel == toRemove {
			delete(e.editors, name)
			return name, toRemove
		}
	}
	return "", nil
}

func (e *TabbedEditor) CurrentEditor() *CodeEditor {
	if e.SelectedPanel() == nil {
		return nil
	}
	return e.SelectedPanel().(*CodeEditor)
}

func (e *TabbedEditor) CurrentFile() string {
	if e.SelectedPanel() == nil {
		return ""
	}
	return e.SelectedPanel().(*CodeEditor).Filepath()
}
