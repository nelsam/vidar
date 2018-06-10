// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package project

import (
	"fmt"

	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

// ProjectSetter is any element that needs to be informed about the
// project being set.
type ProjectSetter interface {
	SetProject(setting.Project)
}

// A Focuser is used to focus elements.
type Focuser interface {
	For(...focus.Opt) bind.Bindable
}

// A Binder is a type which can bind bindables
type Binder interface {
	Pop() []bind.Bindable
	Execute(bind.Bindable)
}

// An Opt is a type that can be used to modify Open ops.
type Opt func(*Open) error

func Name(name string) Opt {
	return func(o *Open) error {
		o.name = name
		return nil
	}
}

func Project(p setting.Project) Opt {
	return func(o *Open) error {
		o.name = p.Name
		return nil
	}
}

// Open is an op that will open a project.
type Open struct {
	status.General

	proj setting.Project

	name    string
	setters []ProjectSetter
	focuser Focuser
	binder  Binder
	hadFile bool
}

func (o *Open) Name() string {
	return "project-change"
}

func (o *Open) For(opts ...Opt) bind.Bindable {
	newO := &Open{
		General: o.General,
		name:    o.name,
		setters: o.setters,
		focuser: o.focuser,
		binder:  o.binder,
		hadFile: o.hadFile,
	}
	for _, opt := range opts {
		if err := opt(newO); err != nil {
			if len(newO.Warn) != 0 {
				newO.Warn += "; "
			}
			newO.Warn += fmt.Sprintf("failed to apply option: %s", err)
		}
	}
	return newO
}

func (o *Open) Reset() {
	o.General.Clear()
	o.setters = nil
	o.focuser = nil
	o.binder = nil
	o.hadFile = false

	o.loadProject()
}

func (o *Open) loadProject() {
	o.proj = setting.Project{}
	for _, proj := range setting.Projects() {
		if proj.Name == o.name {
			o.proj = proj
			return
		}
	}
	o.Err = fmt.Sprintf("No project by the name of '%s' found", o.name)
}

func (o *Open) Store(e interface{}) bind.Status {
	if o.Err != "" {
		return bind.Errored
	}
	switch src := e.(type) {
	case input.Editor:
		o.hadFile = true
	case Binder:
		o.binder = src
	case Focuser:
		o.focuser = src
	case ProjectSetter:
		o.setters = append(o.setters, src)
	}
	if o.binder == nil || o.focuser == nil || len(o.setters) == 0 {
		return bind.Waiting
	}
	return bind.Executing
}

func (o *Open) Exec() error {
	if o.hadFile {
		o.binder.Pop()
	}
	for _, setter := range o.setters {
		setter.SetProject(o.proj)
	}
	o.binder.Execute(o.focuser.For(focus.SkipUnbind()))
	return nil
}
