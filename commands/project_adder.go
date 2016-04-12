// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/settings"
)

const srcDir = string(filepath.Separator) + "src" + string(filepath.Separator)

type projectPane interface {
	Add(settings.Project)
}

type ProjectAdder struct {
	status gxui.Label

	path   *FSLocator
	name   gxui.TextBox
	gopath *FSLocator
	input  <-chan gxui.Focusable
}

func NewProjectAdder(driver gxui.Driver, theme *basic.Theme) *ProjectAdder {
	projectAdder := new(ProjectAdder)
	projectAdder.Init(driver, theme)
	return projectAdder
}

func (p *ProjectAdder) Init(driver gxui.Driver, theme *basic.Theme) {
	p.status = theme.CreateLabel()
	p.path = NewFSLocator(driver, theme)
	p.name = theme.CreateTextBox()
	p.gopath = NewFSLocator(driver, theme)
}

func (p *ProjectAdder) Name() string {
	return "add-project"
}

func (p *ProjectAdder) Start(control gxui.Control) gxui.Control {
	p.path.loadEditorDir(control)

	input := make(chan gxui.Focusable, 3)
	p.input = input
	input <- p.path
	input <- p.name
	input <- p.gopath
	close(input)

	return p.status
}

func (p *ProjectAdder) Next() gxui.Focusable {
	next := <-p.input
	switch next {
	case p.path:
		p.status.SetText("Project Path:")
	case p.name:
		p.status.SetText(fmt.Sprintf("Name for %s", p.path.Path()))
	case p.gopath:
		startPath := p.path.Path()
		lastSrc := strings.LastIndex(startPath, srcDir)
		if lastSrc != -1 {
			startPath = startPath[:lastSrc] + string(filepath.Separator)
		}
		p.gopath.SetPath(startPath)
		p.status.SetText(fmt.Sprintf("GOPATH for %s", p.name.Text()))
	}
	return next
}

func (p *ProjectAdder) Project() settings.Project {
	return settings.Project{
		Name:   p.name.Text(),
		Path:   p.path.Path(),
		Gopath: p.gopath.Path(),
	}
}

func (p *ProjectAdder) Exec(element interface{}) (executed, consume bool) {
	if projects, ok := element.(projectPane); ok {
		project := p.Project()
		settings.AddProject(project)
		projects.Add(project)
		return true, true
	}
	return false, false
}
