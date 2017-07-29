// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/navigator"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/settings"
)

type ProjectSetter interface {
	SetProject(settings.Project)
}

type PaneDisplayer interface {
	ShowNavPane(gxui.Control)
}

type ProjectOpener struct {
	status.General

	name  gxui.TextBox
	input <-chan gxui.Focusable

	proj settings.Project
	nav  PaneDisplayer
}

func NewProjectOpener(theme gxui.Theme) *ProjectOpener {
	p := &ProjectOpener{}
	p.Theme = theme
	p.name = theme.CreateTextBox()
	return p
}

func (p *ProjectOpener) Name() string {
	return "open-project"
}

func (p *ProjectOpener) Menu() string {
	return "File"
}

func (p *ProjectOpener) Start(gxui.Control) gxui.Control {
	p.nav = nil
	p.proj = settings.Project{}

	p.name.SetText("")
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

func (p *ProjectOpener) BeforeExec(interface{}) {
	for _, proj := range settings.Projects() {
		if proj.Name == p.name.Text() {
			p.proj = proj
			break
		}
	}
}

func (p *ProjectOpener) Exec(element interface{}) bind.Status {
	if p.proj.Name != p.name.Text() {
		p.Err = fmt.Sprintf("No project by the name of %s found", p.name.Text())
		return bind.Failed
	}
	switch src := element.(type) {
	case *navigator.ProjectTree:
		if p.nav == nil {
			p.Warn = "No navigation pane found to bind project tree to"
			return bind.Waiting
		}
		src.SetProject(p.proj)
		p.nav.ShowNavPane(src.Frame())
		return bind.Executing
	case ProjectSetter:
		src.SetProject(p.proj)
		return bind.Executing
	case PaneDisplayer:
		p.nav = src
	}
	return bind.Waiting
}
