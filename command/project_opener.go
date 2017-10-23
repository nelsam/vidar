// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/navigator"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
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

	nav     PaneDisplayer
	tree    *navigator.ProjectTree
	setters []ProjectSetter
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
}

func (p *ProjectOpener) Reset() {
	p.nav = nil
	p.setters = nil
	p.tree = nil

	p.proj = settings.Project{}
	for _, proj := range settings.Projects() {
		if proj.Name == p.name.Text() {
			p.proj = proj
			break
		}
	}
}

func (p *ProjectOpener) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case *navigator.ProjectTree:
		p.tree = src
	case ProjectSetter:
		p.setters = append(p.setters, src)
	case PaneDisplayer:
		p.nav = src
	}
	if p.tree == nil {
		return bind.Waiting
	}
	return bind.Executing
}

func (p *ProjectOpener) Exec() error {
	if p.proj.Name != p.name.Text() {
		p.Err = fmt.Sprintf("No project by the name of %s found", p.name.Text())
		return errors.New(p.Err)
	}
	p.tree.SetProject(p.proj)
	p.showTree()
	for _, setter := range p.setters {
		setter.SetProject(p.proj)
	}
	return nil
}

func (p *ProjectOpener) showTree() {
	if p.nav == nil {
		p.Warn = "No navigation pane found to bind project tree to"
		return
	}
	p.nav.ShowNavPane(p.tree.Frame())
}
