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
	"github.com/nelsam/gxui/math"
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

	path          *fs.Locator
	pathRequested bool
	name          gxui.TextBox
	nextEnv       gxui.TextBox
	env           map[string]string

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
	p.name.SetDesiredWidth(math.MaxSize.W)
	p.nextEnv = theme.CreateTextBox()
	p.nextEnv.SetDesiredWidth(math.MaxSize.W)
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
	p.pathRequested = false
	p.name.SetText("")
	p.nextEnv.SetText("")
	p.env = nil

	return p.status
}

func (p *Add) envKeys() (keys []string) {
	for k := range p.env {
		keys = append(keys, k)
	}
	return keys
}

func (p *Add) currEnvs() string {
	envs := p.envKeys()
	if len(envs) == 0 {
		return ""
	}
	return fmt.Sprintf(" [%s]", strings.Join(envs, ", "))
}

func (p *Add) Next() gxui.Focusable {
	if !p.pathRequested {
		p.pathRequested = true
		p.status.SetText("Project Path:")
		return p.path
	}

	if p.name.Text() == "" {
		p.status.SetText(fmt.Sprintf("Name for %s:", p.path.Path()))
		return p.name
	}

	msg := "Environment variables (override environment with VAR==someValue, append with VAR=someValue):"
	defer func() {
		p.status.SetText(msg)
	}()
	if p.env == nil {
		p.env = make(map[string]string)
		return p.nextEnv
	}

	if p.nextEnv.Text() == "" {
		return nil
	}

	idx := strings.IndexRune(p.nextEnv.Text(), '=')
	if idx == -1 {
		msg += fmt.Sprintf("%s ERR: could not parse %s", p.currEnvs(), p.nextEnv.Text())
		return p.nextEnv
	}
	k := p.nextEnv.Text()[:idx]
	v := p.nextEnv.Text()[idx+1:]
	p.env[k] = v
	p.nextEnv.SetText("")
	msg += p.currEnvs()
	return p.nextEnv
}

func (p *Add) Project() setting.Project {
	if _, err := os.Stat(p.path.Path()); os.IsNotExist(err) {
		os.MkdirAll(p.path.Path(), os.ModePerm)
	}
	return setting.Project{
		Name: p.name.Text(),
		Path: p.path.Path(),
		Env:  p.env,
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
	if !p.isCorrectPath(proj.Path) {
		return fmt.Errorf("You can't choose file %s as path for project", proj.Path)
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

func (p *Add) isCorrectPath(path string) bool {
	finfo, err := os.Stat(path)
	if err == nil {
		return finfo.IsDir() //exist dir
	}
	return os.IsNotExist(err) //path not exist
}
