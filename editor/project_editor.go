// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"go/token"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/settings"
)

type ProjectEditor struct {
	SplitEditor

	project settings.Project
}

func NewProjectEditor(driver gxui.Driver, window gxui.Window, theme *basic.Theme, font gxui.Font, project settings.Project) *ProjectEditor {
	p := &ProjectEditor{}
	p.driver = driver
	p.window = window
	p.theme = theme
	p.font = font
	p.SplitterLayout.Init(p, theme)
	p.SetOrientation(gxui.Horizontal)
	p.driver = driver
	p.theme = theme
	p.project = project
	p.SetMouseEventTarget(true)

	p.AddChild(NewTabbedEditor(driver, theme, font))
	return p
}

func (p *ProjectEditor) Open(path string, cursor token.Position) {
	editor := p.open(path)
	p.driver.Call(func() {
		editor.Controller().SetCaret(cursor.Offset)
		editor.ScrollToRune(cursor.Offset)
	})
}

func (p *ProjectEditor) OpenLine(path string, line, col int) {
	editor := p.open(path)
	p.driver.Call(func() {
		lineOffset := editor.LineStart(line)
		editor.Controller().SetCaret(lineOffset + col)
		editor.ScrollToLine(line)
	})
}

func (p *ProjectEditor) open(path string) *CodeEditor {
	return p.SplitEditor.Open(p.project.Path, path, p.project.Gopath, p.project.LicenseHeader())
}

func (p *ProjectEditor) Project() settings.Project {
	return p.project
}

type MultiProjectEditor struct {
	mixins.LinearLayout

	driver gxui.Driver
	theme  *basic.Theme
	font   gxui.Font
	window gxui.Window

	current  *ProjectEditor
	projects map[string]*ProjectEditor
}

func New(driver gxui.Driver, window gxui.Window, theme *basic.Theme, font gxui.Font) *MultiProjectEditor {
	defaultEditor := NewProjectEditor(driver, window, theme, font, settings.DefaultProject)

	e := &MultiProjectEditor{
		projects: map[string]*ProjectEditor{
			"*default*": defaultEditor,
		},
		driver: driver,
		window: window,
		font:   font,
		theme:  theme,
	}
	e.LinearLayout.Init(e, theme)
	e.AddChild(defaultEditor)
	e.current = defaultEditor
	return e
}

func (e *MultiProjectEditor) SetProject(project settings.Project) {
	editor, ok := e.projects[project.Name]
	if !ok {
		editor = NewProjectEditor(e.driver, e.window, e.theme, e.font, project)
		e.projects[project.Name] = editor
	}
	e.RemoveChild(e.current)
	e.AddChild(editor)
	e.current = editor
}

func (e *MultiProjectEditor) CurrentFile() string {
	return e.current.CurrentFile()
}

func (e *MultiProjectEditor) CurrentProject() settings.Project {
	return e.current.Project()
}

func (e *MultiProjectEditor) Focus() {
	e.current.Focus()
}

func (e *MultiProjectEditor) Open(file string, cursor token.Position) {
	e.current.Open(file, cursor)
}
