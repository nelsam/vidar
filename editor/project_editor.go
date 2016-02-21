// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package editor

import (
	"fmt"
	"os"
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

func (p *ProjectEditor) Open(path string, cursor int) {
	name := strings.TrimPrefix(strings.TrimPrefix(path, p.project.Path), "/")
	p.TabbedEditor.New(name, path)
	p.driver.Call(func() {
		p.CurrentEditor().Controller().SetCaret(cursor)
		p.CurrentEditor().ScrollToRune(cursor)
	})
}

func (p *ProjectEditor) Attach() {
	p.TabbedEditor.Attach()
	if p.project.Gopath != "" {
		os.Setenv("GOPATH", p.project.Gopath)
		p.origPath = os.Getenv("PATH")
		os.Setenv("PATH", fmt.Sprintf("%s/bin:%s", p.project.Gopath, p.origPath))
	}
}

func (p *ProjectEditor) Detach() {
	p.TabbedEditor.Detach()
	if p.project.Gopath != "" {
		os.Unsetenv("GOPATH")
		os.Setenv("PATH", p.origPath)
		p.origPath = ""
	}
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

func (e *MultiProjectEditor) Focus() {
	e.current.Focus()
}

func (e *MultiProjectEditor) Open(file string, cursor int) {
	e.current.Open(file, cursor)
}
