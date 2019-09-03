// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/command/fs"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

const srcDir = string(filepath.Separator) + "src" + string(filepath.Separator)

type Adder interface {
	Add(setting.Project)
}

type Add struct {
	status.General

	status gxui.Label

	path   *fs.Locator
	name   gxui.TextBox
	gopath *fs.Locator
	input  <-chan gxui.Focusable

	exec   Executor
	open   *Open
	adders []Adder
}

func NewAdd(driver gxui.Driver, theme *basic.Theme) *Add {
	add := &Add{}
	add.Init(driver, theme)
	return add
}

func (p *Add) Init(driver gxui.Driver, theme *basic.Theme) {
	p.Theme = theme
	p.status = theme.CreateLabel()
	p.path = fs.NewLocator(driver, theme, fs.Dirs)
	p.name = theme.CreateTextBox()
	p.gopath = fs.NewLocator(driver, theme, fs.Dirs)
}

func (p *Add) Name() string {
	return "add-project"
}

func (p *Add) Menu() string {
	return "File"
}

func (p *Add) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyN,
	}}
}

func (p *Add) Start(control gxui.Control) gxui.Control {
	p.path.LoadDir(control)
	p.name.SetText("")

	input := make(chan gxui.Focusable, 3)
	p.input = input
	input <- p.path
	input <- p.name
	input <- p.gopath
	close(input)

	return p.status
}

func (p *Add) Next() gxui.Focusable {
	next := <-p.input
	switch next {
	case p.path:
		p.status.SetText("Project Path:")
	case p.name:
		p.status.SetText(fmt.Sprintf("Name for %s:", p.path.Path()))
	case p.gopath:
		startPath := p.path.Path()
		lastSrc := strings.LastIndex(startPath, srcDir)
		if lastSrc != -1 {
			startPath = startPath[:lastSrc] + string(filepath.Separator)
		}
		p.gopath.SetPath(startPath)
		p.status.SetText(fmt.Sprintf("GOPATH for %s:", p.name.Text()))
	}
	return next
}

func (p *Add) Project() setting.Project {
	if _, err := os.Stat(p.path.Path()); os.IsNotExist(err) {
		os.MkdirAll(p.path.Path(), os.ModePerm)
	}
	return setting.Project{
		Name:   p.name.Text(),
		Path:   p.path.Path(),
		Gopath: p.gopath.Path(),
	}
}

func (p *Add) Reset() {
	p.open = nil
	p.exec = nil
	p.adders = nil
}

func (p *Add) Store(e interface{}) bind.Status {
	switch src := e.(type) {
	case *Open:
		p.open = src
	case Executor:
		p.exec = src
	case Adder:
		p.adders = append(p.adders, src)
	}
	if p.open != nil && p.exec != nil && len(p.adders) > 0 {
		return bind.Executing
	}
	return bind.Waiting
}

func (p *Add) Exec() error {
	proj := p.Project()
	if finfo, err := os.Stat(proj.Path); (err == nil && !finfo.IsDir()) || (err != nil && !os.IsNotExist(err)) {
		return fmt.Errorf("You can't choose file %s as path for project", proj.Path)
	}
	if finfo, err := os.Stat(proj.Gopath); (err == nil && !finfo.IsDir()) || (err != nil && !os.IsNotExist(err)) {
		return fmt.Errorf("You can't choose file %s as gopath for project", proj.Gopath)
	}
	for _, prevProject := range setting.Projects() {
		if prevProject.Name == proj.Name {
			// TODO: Let the user choose a new name
			return fmt.Errorf("There is already a project named %s", proj.Name)
		}
	}
	setting.AddProject(proj)
	for _, adder := range p.adders {
		adder.Add(proj)
	}
	p.exec.Execute(p.open.For(Project(proj)))
	return nil
}
