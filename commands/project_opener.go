// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/settings"
)

type Nav interface {
	ShowNavPane(gxui.Control)
}

type projectSetter interface {
	SetProject(settings.Project)
}

type ProjectOpener struct {
	name     gxui.TextBox
	input    <-chan gxui.Focusable
	projPane gxui.Control
}

func NewProjectOpener(theme gxui.Theme, projPane gxui.Control) *ProjectOpener {
	return &ProjectOpener{
		name:     theme.CreateTextBox(),
		projPane: projPane,
	}
}

func (p *ProjectOpener) Name() string {
	return "open-project"
}

func (p *ProjectOpener) Start(gxui.Control) gxui.Control {
	input := make(chan gxui.Focusable, 1)
	p.input = input
	input <- p.name
	close(input)
	return nil
}

func (p *ProjectOpener) Next() gxui.Focusable {
	return <-p.input
}

func (p *ProjectOpener) SetProject(proj settings.Project) {
	p.name.SetText(proj.Name)
}

func (p *ProjectOpener) BeforeExec(element interface{}) {
	for _, child := range element.(gxui.Parent).Children() {
		if nav, ok := child.Control.(Nav); ok {
			nav.ShowNavPane(p.projPane)
		}
	}
}

func (p *ProjectOpener) Exec(element interface{}) (executed, consume bool) {
	setter, ok := element.(projectSetter)
	if !ok {
		return false, false
	}
	var proj settings.Project
	for _, proj = range settings.Projects() {
		if proj.Name == p.name.Text() {
			break
		}
	}
	if proj.Name != p.name.Text() {
		return false, false
	}
	setter.SetProject(proj)
	return true, false
}
