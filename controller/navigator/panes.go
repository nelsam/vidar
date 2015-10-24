// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package navigator

import (
	"bytes"
	"image"
	"image/draw"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/assets"
	"github.com/nelsam/vidar/settings"
	"github.com/nfnt/resize"

	// Supported image types
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var (
	dirColor = gxui.Color{
		R: 0.8,
		G: 1,
		B: 0.7,
		A: 1.0,
	}
)

type directory struct {
	gxui.AdapterBase
	path    string
	subDirs []string
}

func directories(dirPath string) *directory {
	dir := &directory{
		path: dirPath,
	}
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && path != dirPath && info.IsDir() {
			if !strings.HasPrefix(info.Name(), ".") {
				dir.subDirs = append(dir.subDirs, path)
			}
			return filepath.SkipDir
		}
		return nil
	})
	return dir
}

func (d *directory) Count() int {
	return len(d.subDirs)
}

func (d *directory) NodeAt(index int) gxui.TreeNode {
	return directories(d.subDirs[index])
}

func (d *directory) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	for i, subDir := range d.subDirs {
		if strings.HasPrefix(path, subDir) {
			return i
		}
	}
	return -1
}

func (d *directory) Item() gxui.AdapterItem {
	return d.path
}

func (d *directory) Size(theme gxui.Theme) math.Size {
	return math.Size{
		W: 20 * theme.DefaultMonospaceFont().GlyphMaxSize().W,
		H: theme.DefaultMonospaceFont().GlyphMaxSize().H,
	}
}

func (d *directory) Create(theme gxui.Theme) gxui.Control {
	label := theme.CreateLabel()
	label.SetText(path.Base(d.path))
	label.SetColor(dirColor)
	return label
}

type Directories struct {
	button  gxui.Button
	tree    gxui.Tree
	dirTree *directory
}

func (d *Directories) Init(driver gxui.Driver, theme *basic.Theme) {
	d.button = createIconButton(driver, theme, "folder.png")
	d.tree = theme.CreateTree()
	d.dirTree = directories(os.Getenv("HOME"))
	d.tree.SetAdapter(d.dirTree)
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

func (p *Projects) Init(driver gxui.Driver, theme *basic.Theme) {
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

func (p *Projects) Frame() gxui.Focusable {
	return p.projects
}

func createIconButton(driver gxui.Driver, theme *basic.Theme, iconPath string) gxui.Button {
	button := theme.CreateButton()
	button.SetType(gxui.PushButton)

	fileBytes, err := assets.Asset(iconPath)
	if err != nil {
		panic(err)
	}
	f := bytes.NewBuffer(fileBytes)
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
