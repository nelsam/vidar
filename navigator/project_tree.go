// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"go/token"
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
	Position() token.Position
}

type ProjectTree struct {
	button gxui.Button

	driver gxui.Driver
	theme  *basic.Theme

	dirs        *dirTree
	dirsAdapter *dirTreeAdapter

	project        *dirTree
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

func (p *ProjectTree) SetHeight(height int) {
	p.project.height = height - p.dirs.height
	p.project.SizeChanged()
}

func (p *ProjectTree) Button() gxui.Button {
	return p.button
}

func (p *ProjectTree) SetRoot(path string) {
	p.dirsAdapter = &dirTreeAdapter{}
	p.dirsAdapter.children = []string{path}
	p.dirsAdapter.dirs = true
	p.dirs.SetAdapter(p.dirsAdapter)

	p.projectAdapter = NewTOC(path)
	p.project.SetAdapter(p.projectAdapter)

	p.project.ExpandAll()
}

func (p *ProjectTree) SetProject(project settings.Project) {
	p.SetRoot(project.Path)
}

func (p *ProjectTree) Open(filePath string) {
	dir, _ := filepath.Split(filePath)
	p.dirs.Select(dir)
}

func (p *ProjectTree) Frame() gxui.Control {
	return p.layout
}

func (p *ProjectTree) OnComplete(onComplete func(controller.Executor)) {
	cmd := commands.NewFileOpener(p.driver, p.theme)
	p.project.OnSelectionChanged(func(selected gxui.AdapterItem) {
		var node gxui.TreeNode = p.projectAdapter
		for i := node.ItemIndex(selected); i != -1; i = node.ItemIndex(selected) {
			node = node.NodeAt(i)
		}
		locationer, ok := node.(Locationer)
		if !ok {
			return
		}
		cmd.SetLocation(locationer.File(), locationer.Position())
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
	theme  *basic.Theme
	height int
}

func newDirTree(theme *basic.Theme) *dirTree {
	t := &dirTree{
		theme:  theme,
		height: 20 * theme.DefaultMonospaceFont().GlyphMaxSize().H,
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
	height := t.height
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
