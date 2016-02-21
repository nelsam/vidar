package navigator

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/controller"
	"github.com/nelsam/vidar/settings"
)

var (
	dirColor = gxui.Color{
		R: 0.8,
		G: 1,
		B: 0.7,
		A: 1,
	}
	fileColor = gxui.Gray80

	splitterBarBackgroundColor = gxui.Color{
		R: 0.6,
		G: 0.3,
		B: 0.3,
		A: 1,
	}
	splitterBarForegroundColor = gxui.Color{
		R: 0.8,
		G: 0.4,
		B: 0.4,
		A: 1,
	}
)

type Locationer interface {
	File() string
	Pos() int
}

type ProjectTree struct {
	button gxui.Button

	driver gxui.Driver
	theme  *basic.Theme

	dirs        gxui.Tree
	dirsAdapter *dirTreeAdapter

	project        gxui.Tree
	projectAdapter *TOC

	layout gxui.LinearLayout
}

func NewProjectTree(driver gxui.Driver, theme *basic.Theme) *ProjectTree {
	tree := &ProjectTree{
		driver:  driver,
		theme:   theme,
		button:  createIconButton(driver, theme, "folder.png"),
		dirs:    newDirTree(theme),
		project: newDirTree(theme),
		layout:  theme.CreateLinearLayout(),
	}
	tree.layout.SetDirection(gxui.TopToBottom)
	tree.layout.AddChild(tree.dirs)
	tree.layout.AddChild(tree.project)

	tree.SetProject(settings.DefaultProject)

	tree.dirs.OnSelectionChanged(func(selection gxui.AdapterItem) {
		tree.dirs.Show(selection)
		tree.projectAdapter = NewTOC(selection.(string))
		tree.project.SetAdapter(tree.projectAdapter)
		tree.project.ExpandAll()
	})

	return tree
}

func (d *ProjectTree) Button() gxui.Button {
	return d.button
}

func (d *ProjectTree) SetRoot(path string) {
	log.Printf("Setting the dir adapter")
	d.dirsAdapter = &dirTreeAdapter{}
	d.dirsAdapter.children = []string{path}
	d.dirsAdapter.dirs = true
	d.dirs.SetAdapter(d.dirsAdapter)

	log.Printf("Setting the toc")
	d.projectAdapter = NewTOC(path)
	d.project.SetAdapter(d.projectAdapter)

	d.project.ExpandAll()
	log.Printf("Done setting the project")
}

func (d *ProjectTree) SetProject(project settings.Project) {
	log.Printf("SetProject called with project %#v", project)
	d.SetRoot(project.Path)
}

func (d *ProjectTree) Open(filePath string) {
	dir, _ := filepath.Split(filePath)
	d.dirs.Select(dir)
}

func (d *ProjectTree) Frame() gxui.Control {
	return d.layout
}

func (d *ProjectTree) OnComplete(onComplete func(controller.Executor)) {
	cmd := commands.NewFileOpener(d.driver, d.theme)
	d.project.OnSelectionChanged(func(selected gxui.AdapterItem) {
		var node gxui.TreeNode = d.projectAdapter
		for i := node.ItemIndex(selected); i != -1; i = node.ItemIndex(selected) {
			node = node.NodeAt(i)
		}
		locationer, ok := node.(Locationer)
		if !ok {
			return
		}
		cmd.SetLocation(locationer.File(), locationer.Pos())
		onComplete(cmd)
	})
}

type fsAdapter struct {
	gxui.AdapterBase
	path     string
	children []string
	dirs     bool
}

func fsList(dirPath string, dirs bool) fsAdapter {
	fs := fsAdapter{
		path: dirPath,
		dirs: dirs,
	}
	fs.Walk()
	return fs
}

func (a *fsAdapter) Walk() {
	filepath.Walk(a.path, func(path string, info os.FileInfo, err error) error {
		if err == nil && path != a.path {
			use := (a.dirs && info.IsDir()) || (!a.dirs && !info.IsDir())
			if use && !strings.HasPrefix(info.Name(), ".") {
				a.children = append(a.children, path)
			}
			if info.IsDir() {
				return filepath.SkipDir
			}
		}
		return nil
	})
}

func (a *fsAdapter) Count() int {
	return len(a.children)
}

func (a *fsAdapter) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	for i, subDir := range a.children {
		if strings.HasPrefix(path, subDir) {
			return i
		}
	}
	return -1
}

func (a *fsAdapter) Size(theme gxui.Theme) math.Size {
	return math.Size{
		W: 20 * theme.DefaultMonospaceFont().GlyphMaxSize().W,
		H: theme.DefaultMonospaceFont().GlyphMaxSize().H,
	}
}

func (a *fsAdapter) create(theme gxui.Theme, path string) gxui.Label {
	label := theme.CreateLabel()
	label.SetText(filepath.Base(path))
	return label
}

type dirTree struct {
	mixins.Tree
	theme *basic.Theme
}

func newDirTree(theme *basic.Theme) gxui.Tree {
	t := &dirTree{
		theme: theme,
	}
	t.Init(t, theme)
	return t
}

func (t *dirTree) DesiredSize(min, max math.Size) math.Size {
	width := 20 * t.theme.DefaultMonospaceFont().GlyphMaxSize().W
	if min.W > width {
		width = min.W
	}
	if max.W < width {
		width = max.W
	}
	height := 20 * t.theme.DefaultMonospaceFont().GlyphMaxSize().H
	if min.H > height {
		height = min.H
	}
	if max.H < height {
		height = max.H
	}
	size := math.Size{
		W: width,
		H: height,
	}
	return size
}

type dirTreeAdapter struct {
	fsAdapter
}

func loadDirTreeAdapter(dirPath string) *dirTreeAdapter {
	return &dirTreeAdapter{
		fsAdapter: fsList(dirPath, true),
	}
}

func (a *dirTreeAdapter) Item() gxui.AdapterItem {
	return a.path
}

func (a *dirTreeAdapter) NodeAt(index int) gxui.TreeNode {
	return loadDirTreeAdapter(a.children[index])
}

func (a *dirTreeAdapter) Create(theme gxui.Theme) gxui.Control {
	label := a.create(theme, a.path)
	label.SetColor(dirColor)
	return label
}

type fileListAdapter struct {
	fsAdapter
}

func fileList(dirPath string) *fileListAdapter {
	return &fileListAdapter{
		fsAdapter: fsList(dirPath, false),
	}
}

func (a *fileListAdapter) ItemAt(index int) gxui.AdapterItem {
	return a.children[index]
}

func (a *fileListAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	label := a.create(theme, a.children[index])
	label.SetColor(fileColor)
	return label
}
