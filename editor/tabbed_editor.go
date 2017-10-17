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
	"github.com/nelsam/vidar/theme"
)

type TabbedEditor struct {
	mixins.PanelHolder

	editors map[string]*CodeEditor

	driver      gxui.Driver
	cmdr        Commander
	theme       *basic.Theme
	syntaxTheme theme.Theme
	font        gxui.Font
}

func NewTabbedEditor(driver gxui.Driver, cmdr Commander, theme *basic.Theme, syntaxTheme theme.Theme, font gxui.Font) *TabbedEditor {
	editor := &TabbedEditor{}
	editor.Init(editor, driver, cmdr, theme, syntaxTheme, font)
	return editor
}

func (e *TabbedEditor) Init(outer mixins.PanelHolderOuter, driver gxui.Driver, cmdr Commander, theme *basic.Theme, syntaxTheme theme.Theme, font gxui.Font) {
	e.editors = make(map[string]*CodeEditor)
	e.driver = driver
	e.cmdr = cmdr
	e.theme = theme
	e.syntaxTheme = syntaxTheme
	e.font = font
	e.PanelHolder.Init(outer, theme)
	e.SetMargin(math.Spacing{L: 0, T: 2, R: 0, B: 0})
}

func (e *TabbedEditor) Has(hiddenPrefix, path string) bool {
	_, ok := e.editors[relPath(hiddenPrefix, path)]
	return ok
}

func (e *TabbedEditor) Open(hiddenPrefix, path, headerText string, environ []string) (editor *CodeEditor, existed bool) {
	name := relPath(hiddenPrefix, path)
	if editor, ok := e.editors[name]; ok {
		e.Select(e.PanelIndex(editor))
		gxui.SetFocus(editor)
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
			gxui.SetFocus(focused.(gxui.Focusable))
		})
	})
	editor.Init(e.driver, e.theme, e.syntaxTheme, e.font, path, headerText)
	editor.SetTabWidth(4)
	e.Add(name, editor)
	return editor, false
}

func (e *TabbedEditor) Add(name string, editor *CodeEditor) {
	e.editors[name] = editor
	e.AddPanel(editor, name)
	e.Select(e.PanelIndex(editor))
	gxui.SetFocus(editor)
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
		opener := e.cmdr.Bindable("open-file").(Opener)
		e.cmdr.Execute(opener.For(e.CurrentEditor().Filepath(), -1))
	})
	return tab
}

func (e *TabbedEditor) EditorAt(d Direction) *CodeEditor {
	panels := e.PanelCount()
	if panels < 2 {
		return e.CurrentEditor()
	}
	idx := e.PanelIndex(e.SelectedPanel())
	switch d {
	case Right:
		idx++
		if idx == panels {
			idx = 0
		}
	case Left:
		idx--
		if idx < 0 {
			idx = panels - 1
		}
	}
	return e.Panel(idx).(*CodeEditor)
}

func (e *TabbedEditor) CloseCurrentEditor() (name string, editor *CodeEditor) {
	toRemove := e.CurrentEditor()
	if toRemove == nil {
		return "", nil
	}
	e.RemovePanel(toRemove)
	defer func() {
		if ed := e.CurrentEditor(); ed != nil {
			opener := e.cmdr.Bindable("open-file").(Opener)
			e.cmdr.Execute(opener.For(ed.Filepath(), -1))
		}
	}()
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

func (e *TabbedEditor) Elements() []interface{} {
	if e.SelectedPanel() == nil {
		return nil
	}
	return []interface{}{e.SelectedPanel()}
}

func relPath(from, path string) string {
	rel := strings.TrimPrefix(path, from)
	if rel[0] == filepath.Separator {
		rel = rel[1:]
	}
	return rel
}
