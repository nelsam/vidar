// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"log"
	"os"
	"path/filepath"
	"strings"

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

func (e *TabbedEditor) Open(hiddenPrefix, path, headerText string, environ []string) (editor *CodeEditor, existed bool) {
	name := relPath(hiddenPrefix, path)
	if editor, ok := e.editors[name]; ok {
		e.Select(e.PanelIndex(editor))
		e.Focus()
		return editor, true
	}
	editor = &CodeEditor{}
	// We want the OnRename trigger set up before the editor opens the file
	// in its Init method.
	editor.OnRename(func(newPath string) {
		e.driver.Call(func() {
			delete(e.editors, name)
			newName := relPath(hiddenPrefix, newPath)
			focused := e.SelectedPanel()
			e.editors[newName] = editor
			idx := e.PanelIndex(editor)
			if idx == -1 {
				return
			}
			e.RemovePanel(editor)
			e.AddPanelAt(editor, newName, idx)
			e.Select(e.PanelIndex(focused))
			e.Focus()
		})
	})
	editor.Init(e.driver, e.theme, e.font, path, headerText)
	editor.SetTabWidth(4)
	suggester := suggestions.NewGoCodeProvider(editor, environ)
	editor.SetSuggestionProvider(suggester)
	e.Add(name, editor)
	return editor, false
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
	tab := basic.CreatePanelTab(e.theme)
	tab.OnMouseUp(func(gxui.MouseEvent) {
		e.driver.Call(e.Focus)
	})
	return tab
}

func (e *TabbedEditor) ShiftTab(delta int) {
	panels := e.PanelCount()
	if panels < 2 {
		return
	}
	current := e.PanelIndex(e.SelectedPanel())
	next := current + delta
	for next < 0 {
		next = panels + next
	}
	next = next % panels
	e.Select(next)
	e.driver.Call(func() {
		e.Focus()
	})
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

func (e *TabbedEditor) SaveAll() {
	for name, editor := range e.editors {
		f, err := os.Create(name)
		if err != nil {
			log.Printf("Could not save %s : %s", name, err)
		}
		defer f.Close()
		if _, err := f.WriteString(editor.Text()); err != nil {
			log.Printf("Could not write to file %s: %s", name, err)
		}
	}
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

func relPath(from, path string) string {
	rel := strings.TrimPrefix(path, from)
	if rel[0] == filepath.Separator {
		rel = rel[1:]
	}
	return rel
}
