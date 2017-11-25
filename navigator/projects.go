// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/setting"
)

type ProjectSetter interface {
	bind.Bindable

	SetProject(setting.Project)
}

type Projects struct {
	theme gxui.Theme
	cmdr  Commander

	button          gxui.Button
	projects        gxui.List
	projectsAdapter *gxui.DefaultAdapter

	projectFrame gxui.Control
}

func NewProjectsPane(cmdr Commander, driver gxui.Driver, theme gxui.Theme, projFrame gxui.Control) *Projects {
	pane := &Projects{
		cmdr:            cmdr,
		theme:           theme,
		projectFrame:    projFrame,
		button:          createIconButton(driver, theme, "projects.png"),
		projects:        theme.CreateList(),
		projectsAdapter: gxui.CreateDefaultAdapter(),
	}
	pane.projectsAdapter.SetItems(setting.Projects())
	pane.projects.SetAdapter(pane.projectsAdapter)
	pane.projects.OnSelectionChanged(func(selected gxui.AdapterItem) {
		opener := pane.cmdr.Bindable("open-project").(ProjectSetter)
		opener.SetProject(selected.(setting.Project))
		pane.cmdr.Execute(opener)
	})
	return pane
}

func (p *Projects) Add(project setting.Project) {
	projects := append(p.projectsAdapter.Items().([]setting.Project), project)
	p.projectsAdapter.SetItems(projects)
}

func (p *Projects) Button() gxui.Button {
	return p.button
}

func (p *Projects) Frame() gxui.Control {
	return p.projects
}

func (p *Projects) Projects() []setting.Project {
	return p.projectsAdapter.Items().([]setting.Project)
}
