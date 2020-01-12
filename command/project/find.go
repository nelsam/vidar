// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package project

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

// An Executor can execute bindables
type Executor interface {
	Execute(bind.Bindable)
}

type Find struct {
	status.General

	name  gxui.TextBox
	input <-chan gxui.Focusable

	proj setting.Project

	exec Executor
	open *Open
}

func NewFind(theme gxui.Theme) *Find {
	p := &Find{}
	p.Theme = theme
	p.name = theme.CreateTextBox()
	p.name.SetDesiredWidth(math.MaxSize.W)
	return p
}

func (p *Find) Name() string {
	return "open-project"
}

func (p *Find) Menu() string {
	return "File"
}

func (p *Find) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyO,
	}}
}

func (p *Find) Start(gxui.Control) gxui.Control {
	p.name.SetText("")
	input := make(chan gxui.Focusable, 1)
	input <- p.name
	p.input = input
	close(input)
	return nil
}

func (p *Find) Next() gxui.Focusable {
	return <-p.input
}

func (p *Find) Reset() {
	p.exec = nil
	p.open = nil
}

func (p *Find) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case *Open:
		p.open = src
	case Executor:
		p.exec = src
	}
	if p.exec == nil || p.open == nil {
		return bind.Waiting
	}
	return bind.Executing
}

func (p *Find) Exec() error {
	p.exec.Execute(p.open.For(Name(p.name.Text())))
	return nil
}
