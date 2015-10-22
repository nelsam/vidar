// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package navigator

import (
	// Supported image types
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/settings"
	"github.com/nfnt/resize"
)

type fakeNode struct {
	gxui.AdapterBase
	path     string
	subPaths []string
}

func (f fakeNode) Count() int {
	return len(f.subPaths)
}

func (f fakeNode) NodeAt(index int) gxui.TreeNode {
	return fakeNode{
		path: path.Join(f.path, f.subPaths[index]),
	}
}

func (f fakeNode) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	for i, subPath := range f.subPaths {
		if strings.HasPrefix(path, subPath) {
			return i
		}
	}
	return -1
}

func (f fakeNode) Item() gxui.AdapterItem {
	return f.path
}

func (f fakeNode) Size(theme gxui.Theme) math.Size {
	return math.Size{
		W: 20 * theme.DefaultMonospaceFont().GlyphMaxSize().W,
		H: math.MaxSize.H,
	}
}

func (f fakeNode) Create(theme gxui.Theme) gxui.Control {
	label := theme.CreateLabel()
	label.SetText(f.path)
	return label
}

type Directories struct {
	button gxui.Button
	tree   gxui.Tree
}

func (d *Directories) Init(driver gxui.Driver, theme *basic.Theme, assetsDir string) {
	d.button = createIconButton(driver, theme, path.Join(assetsDir, "folder.png"))
	d.tree = theme.CreateTree()
	d.tree.SetAdapter(&fakeNode{
		path: "foo",
		subPaths: []string{
			"bar",
			"baz",
		},
	})
}

func (d *Directories) Button() gxui.Button {
	return d.button
}

func (d *Directories) Frame() gxui.Focusable {
	return d.tree
}

type Projects struct {
	button          gxui.Button
	projects        gxui.List
	projectsAdapter *gxui.DefaultAdapter
}

func (p *Projects) Init(driver gxui.Driver, theme *basic.Theme, assetsDir string) {
	p.button = createIconButton(driver, theme, path.Join(assetsDir, "projects.png"))
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

func (p *Projects) Frame() gxui.Focusable {
	return p.projects
}

func createIconButton(driver gxui.Driver, theme *basic.Theme, iconPath string) gxui.Button {
	button := theme.CreateButton()
	button.SetType(gxui.PushButton)

	f, err := os.Open(iconPath)
	if err != nil {
		panic(err)
	}
	src, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	src = resize.Resize(24, 24, src, resize.Bilinear)

	rgba := image.NewRGBA(src.Bounds())
	draw.Draw(rgba, src.Bounds(), src, image.ZP, draw.Src)
	texture := driver.CreateTexture(rgba, 1)

	icon := theme.CreateImage()
	icon.SetTexture(texture)
	button.AddChild(icon)
	return button
}
