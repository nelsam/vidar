// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
)

type projectSetter interface {
	SetProject(string)
}

type ProjectOpener struct {
	name  gxui.TextBox
	input <-chan gxui.Focusable
}

func NewProjectOpener(driver gxui.Driver, theme *basic.Theme) *ProjectOpener {
	projectOpener := new(ProjectOpener)
	projectOpener.Init(driver, theme)
	return projectOpener
}

func (p *ProjectOpener) Init(driver gxui.Driver, theme *basic.Theme) {
	p.name = theme.CreateTextBox()
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

func (p *ProjectOpener) SetProject(projName string) {
	p.name.SetText(projName)
}

func (p *ProjectOpener) Exec(element interface{}) (executed, consume bool) {
	setter, ok := element.(projectSetter)
	if !ok {
		return false, false
	}
	setter.SetProject(p.name.Text())
	return true, false
}
