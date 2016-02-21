// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package navigator

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/controller"
	"github.com/nelsam/vidar/settings"
)

type Projects struct {
	driver gxui.Driver
	theme  *basic.Theme

	button          gxui.Button
	projects        gxui.List
	projectsAdapter *gxui.DefaultAdapter
}

func (p *Projects) Init(driver gxui.Driver, theme *basic.Theme) {
	p.driver = driver
	p.theme = theme
	p.button = createIconButton(driver, theme, "projects.png")
	p.projects = theme.CreateList()
	p.projectsAdapter = gxui.CreateDefaultAdapter()

	p.projectsAdapter.SetItems(settings.Projects())
	p.projects.SetAdapter(p.projectsAdapter)
}

func (p *Projects) Add(project settings.Project) {
	projects := append(p.projectsAdapter.Items().([]settings.Project), project)
	p.projectsAdapter.SetItems(projects)
}

func (p *Projects) Button() gxui.Button {
	return p.button
}

func (p *Projects) Frame() gxui.Control {
	return p.projects
}

func (p *Projects) Projects() []settings.Project {
	return p.projectsAdapter.Items().([]settings.Project)
}

func (p *Projects) OnComplete(onComplete func(controller.Executor)) {
	opener := commands.NewProjectOpener(p.driver, p.theme)
	p.projects.OnSelectionChanged(func(selected gxui.AdapterItem) {
		opener.SetProject(selected.(settings.Project))
		onComplete(opener)
	})
}
