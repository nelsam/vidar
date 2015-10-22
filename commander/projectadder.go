// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui_playground/controller"
	"github.com/nelsam/gxui_playground/settings"
)

type projectPane interface {
	Add(settings.Project)
}

type ProjectAdder struct {
	path  *FSLocator
	name  gxui.TextBox
	items <-chan gxui.Focusable
}

func NewProjectAdder(driver gxui.Driver, theme *basic.Theme) controller.Command {
	projectAdder := new(ProjectAdder)
	projectAdder.Init(driver, theme)
	return projectAdder
}

func (p *ProjectAdder) Init(driver gxui.Driver, theme *basic.Theme) {
	p.path = NewFSLocator(driver, theme)
	p.name = theme.CreateTextBox()
}

func (p *ProjectAdder) Name() string {
	return "add-project"
}

func (p *ProjectAdder) Start(control gxui.Control) gxui.Control {
	p.path.loadEditorDir(control)

	items := make(chan gxui.Focusable, 2)
	items <- p.path
	items <- p.name
	close(items)
	p.items = items

	return nil
}

func (p *ProjectAdder) Next() gxui.Focusable {
	return <-p.items
}

func (p *ProjectAdder) Project() settings.Project {
	return settings.Project{
		Name: p.name.Text(),
		Path: p.path.Path(),
	}
}

func (p *ProjectAdder) Exec(element interface{}) (consume bool) {
	if projects, ok := element.(projectPane); ok {
		project := p.Project()
		settings.AddProject(project)
		projects.Add(project)
		return true
	}
	return false
}
