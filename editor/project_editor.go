// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package editor

import (
	"go/token"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/settings"
)

type ProjectEditor struct {
	TabbedEditor

	project  settings.Project
	origPath string
}

func (p *ProjectEditor) Init(driver gxui.Driver, theme *basic.Theme, font gxui.Font, project settings.Project) {
	p.TabbedEditor.Init(p, driver, theme, font)
	p.project = project
}

func (p *ProjectEditor) Open(path string, cursor token.Position) {
	name := strings.TrimPrefix(strings.TrimPrefix(path, p.project.Path), "/")
	editor := p.TabbedEditor.New(name, path, p.project.Gopath)
	p.driver.Call(func() {
		editor.Controller().SetCaret(cursor.Offset)
		editor.ScrollToRune(cursor.Offset)
	})
}

func (p *ProjectEditor) Attach() {
	p.TabbedEditor.Attach()
}

func (p *ProjectEditor) Detach() {
	p.TabbedEditor.Detach()
}

func (p *ProjectEditor) Project() settings.Project {
	return p.project
}

type MultiProjectEditor struct {
	mixins.LinearLayout

	driver gxui.Driver
	theme  *basic.Theme
	font   gxui.Font

	current  *ProjectEditor
	projects map[string]*ProjectEditor
}

func New(driver gxui.Driver, theme *basic.Theme, font gxui.Font) *MultiProjectEditor {
	defaultEditor := new(ProjectEditor)
	defaultEditor.Init(driver, theme, font, settings.DefaultProject)

	e := &MultiProjectEditor{
		projects: map[string]*ProjectEditor{
			"*default*": defaultEditor,
		},
	}
	e.Init(driver, theme, font)
	e.AddChild(defaultEditor)
	e.current = defaultEditor
	return e
}

func (e *MultiProjectEditor) Init(driver gxui.Driver, theme *basic.Theme, font gxui.Font) {
	e.driver = driver
	e.theme = theme
	e.font = font
	e.LinearLayout.Init(e, e.theme)
}

func (e *MultiProjectEditor) SetProject(project settings.Project) {
	editor, ok := e.projects[project.Name]
	if !ok {
		editor = new(ProjectEditor)
		editor.Init(e.driver, e.theme, e.font, project)
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
