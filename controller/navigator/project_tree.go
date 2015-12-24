package navigator

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commands"
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

type ProjectTree struct {
	button gxui.Button

	driver gxui.Driver
	theme  *basic.Theme

	dirs        gxui.Tree
	dirsAdapter *dirTreeAdapter

	files        gxui.List
	filesAdapter *fileListAdapter

	layout gxui.LinearLayout
}

func (d *ProjectTree) Init(driver gxui.Driver, theme *basic.Theme) {
	d.driver = driver
	d.theme = theme

	d.button = createIconButton(driver, theme, "folder.png")
	d.dirs = newDirTree(theme)
	d.files = theme.CreateList()

	d.layout = theme.CreateLinearLayout()
	d.layout.SetDirection(gxui.TopToBottom)
	d.layout.AddChild(d.dirs)
	d.layout.AddChild(d.files)

	d.SetProject(settings.DefaultProject)

	d.dirs.OnSelectionChanged(func(selection gxui.AdapterItem) {
		d.filesAdapter = fileList(selection.(string))
		d.files.SetAdapter(d.filesAdapter)
	})
}

func (d *ProjectTree) Button() gxui.Button {
	return d.button
}

func (d *ProjectTree) SetRoot(path string) {
	d.dirsAdapter = &dirTreeAdapter{}
	d.dirsAdapter.children = []string{path}
	d.dirsAdapter.dirs = true
	d.dirs.SetAdapter(d.dirsAdapter)

	d.filesAdapter = fileList(path)
	d.files.SetAdapter(d.filesAdapter)
}

func (d *ProjectTree) SetProject(project settings.Project) {
	d.SetRoot(project.Path)
}

func (d *ProjectTree) Open(filePath string) {
	dir, file := filepath.Split(filePath)
	d.dirs.Select(dir)
	d.files.Select(file)
}

func (d *ProjectTree) Frame() gxui.Control {
	return d.layout
}

func (d *ProjectTree) OnComplete(onComplete func(commands.Command)) {
	cmd := commands.NewFileOpener(d.driver, d.theme)
	d.files.OnSelectionChanged(func(selected gxui.AdapterItem) {
		cmd.SetPath(selected.(string))
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
