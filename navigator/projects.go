// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/command/project"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/setting"
)

type ProjectChanger interface {
	For(...project.Opt) bind.Bindable
}

type Projects struct {
	theme gxui.Theme
	cmdr  Commander

	button          gxui.Button
	projects        gxui.List
	projectsAdapter *gxui.DefaultAdapter

	projectMap   map[string]setting.Project
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
		projectMap:      make(map[string]setting.Project),
	}
	var names []string
	for _, p := range setting.Projects() {
		names = append(names, p.Name)
		pane.projectMap[p.Name] = p
	}
	pane.projectsAdapter.SetItems(names)
	pane.projects.SetAdapter(pane.projectsAdapter)
	pane.projects.OnSelectionChanged(func(selected gxui.AdapterItem) {
		opener := pane.cmdr.Bindable("project-change").(ProjectChanger)
		pane.cmdr.Execute(opener.For(project.Project(pane.projectMap[selected.(string)])))
	})
	return pane
}

func (p *Projects) Add(project setting.Project) {
	p.projectMap[project.Name] = project
	projects := append(p.projectsAdapter.Items().([]string), project.Name)
	p.projectsAdapter.SetItems(projects)
}

func (p *Projects) Button() gxui.Button {
	return p.button
}

func (p *Projects) Frame() gxui.Control {
	return p.projects
}

func (p *Projects) Projects() []setting.Project {
	var projs []setting.Project
	for _, proj := range p.projectMap {
		projs = append(projs, proj)
	}
	return projs
}
