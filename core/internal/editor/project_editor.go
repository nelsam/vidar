// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/setting"
	"github.com/nelsam/vidar/theme"
	"github.com/nelsam/vidar/ui"
	"github.com/pkg/errors"
)

type deprecatedGXUIDriverTheme interface {
	Raw() (gxui.Driver, gxui.Theme)
}

type deprecatedGXUICoupling interface {
	Control() gxui.Control
}

type ProjectEditor struct {
	SplitEditor

	project setting.Project
}

func NewProjectEditor(creator ui.Creator, window gxui.Window, cmdr Commander, syntaxTheme theme.Theme, project setting.Project) (*ProjectEditor, error) {
	p := &ProjectEditor{}
	d, t := creator.(deprecatedGXUIDriverTheme).Raw()
	p.driver = d
	p.window = window
	p.cmdr = cmdr
	p.theme = t.(*basic.Theme)
	p.syntaxTheme = syntaxTheme
	p.font = t.DefaultMonospaceFont()
	p.SplitterLayout.Init(p, t)
	p.SetOrientation(gxui.Horizontal)
	p.project = project
	p.SetMouseEventTarget(true)

	e, err := NewTabbedEditor(creator, cmdr, syntaxTheme)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tabbed editor")
	}
	p.AddChild(e.layout.(deprecatedGXUICoupling).Control())
	return p, nil
}

func (p *ProjectEditor) Open(path string) (e input.Editor, existed bool) {
	return p.SplitEditor.Open(p.project.Path, path, p.project.LicenseHeader(), p.project.Environ())
}

func (p *ProjectEditor) Project() setting.Project {
	return p.project
}

type MultiProjectEditor struct {
	mixins.LinearLayout

	creator     ui.Creator
	driver      gxui.Driver
	cmdr        Commander
	theme       *basic.Theme
	syntaxTheme theme.Theme
	font        gxui.Font
	window      gxui.Window

	current  *ProjectEditor
	projects map[string]*ProjectEditor
}

func New(creator ui.Creator, window gxui.Window, cmdr Commander, syntaxTheme theme.Theme) (*MultiProjectEditor, error) {
	defaultEditor, err := NewProjectEditor(creator, window, cmdr, syntaxTheme, setting.DefaultProject)
	if err != nil {
		return nil, errors.Wrap(err, "could not create project editor")
	}

	d, t := creator.(deprecatedGXUIDriverTheme).Raw()
	e := &MultiProjectEditor{
		projects: map[string]*ProjectEditor{
			"*default*": defaultEditor,
		},
		creator:     creator,
		driver:      d,
		window:      window,
		cmdr:        cmdr,
		font:        t.DefaultMonospaceFont(),
		theme:       t.(*basic.Theme),
		syntaxTheme: syntaxTheme,
	}
	e.LinearLayout.Init(e, t)
	e.AddChild(defaultEditor)
	e.current = defaultEditor
	return e, nil
}

func (e *MultiProjectEditor) SetProject(project setting.Project) error {
	editor, ok := e.projects[project.Name]
	if !ok {
		var err error
		editor, err = NewProjectEditor(e.creator, e.window, e.cmdr, e.syntaxTheme, project)
		e.projects[project.Name] = editor
		return errors.Wrap(err, "could not create project editor")
	}
	e.RemoveChild(e.current)
	e.AddChild(editor)
	e.current = editor
	return nil
}

func (e *MultiProjectEditor) Elements() []interface{} {
	return []interface{}{
		e.current,
	}
}

func (e *MultiProjectEditor) CurrentEditor() input.Editor {
	return e.current.CurrentEditor()
}

func (e *MultiProjectEditor) CurrentFile() string {
	return e.current.CurrentFile()
}

func (e *MultiProjectEditor) CurrentProject() setting.Project {
	return e.current.Project()
}

func (e *MultiProjectEditor) Open(file string) (ed input.Editor, existed bool) {
	return e.current.Open(file)
}
