// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
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

// A Binder is a type which can bind bindables
type Binder interface {
	Pop() []bind.Bindable
	Execute(bind.Bindable)
}

type ProjectOpener struct {
	status.General

	name  gxui.TextBox
	input <-chan gxui.Focusable

	proj settings.Project

	nav     PaneDisplayer
	focuser Focuser
	binder  Binder
	editor  input.Editor
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

func (p *ProjectOpener) Reset() {
	p.nav = nil
	p.setters = nil
	p.tree = nil
	p.focuser = nil
	p.binder = nil
	p.editor = nil

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
	case input.Editor:
		p.editor = src
	case Binder:
		p.binder = src
	case Focuser:
		p.focuser = src
	case *navigator.ProjectTree:
		p.tree = src
	case ProjectSetter:
		p.setters = append(p.setters, src)
	case PaneDisplayer:
		p.nav = src
	}
	if p.tree == nil || p.binder == nil || p.focuser == nil {
		return bind.Waiting
	}
	return bind.Executing
}

func (p *ProjectOpener) Exec() error {
	if p.proj.Name != p.name.Text() {
		p.Err = fmt.Sprintf("No project by the name of %s found", p.name.Text())
		return errors.New(p.Err)
	}
	if p.editor != nil {
		p.binder.Pop()
	}
	p.tree.SetProject(p.proj)
	p.showTree()
	for _, setter := range p.setters {
		setter.SetProject(p.proj)
	}
	p.binder.Execute(p.focuser.For(focus.SkipUnbind()))
	return nil
}

func (p *ProjectOpener) showTree() {
	if p.nav == nil {
		p.Warn = "No navigation pane found to bind project tree to"
		return
	}
	p.nav.ShowNavPane(p.tree.Frame())
}
