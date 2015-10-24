package navigator

import (
	"fmt"
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

	layout *splitterLayout
}

func (d *ProjectTree) Init(driver gxui.Driver, theme *basic.Theme) {
	d.driver = driver
	d.theme = theme

	d.button = createIconButton(driver, theme, "folder.png")
	d.dirs = theme.CreateTree()
	d.dirsAdapter = dirTree(os.Getenv("HOME"))
	d.dirs.SetAdapter(d.dirsAdapter)

	d.files = theme.CreateList()
	d.filesAdapter = fileList(os.Getenv("HOME"))
	d.files.SetAdapter(d.filesAdapter)

	d.layout = newSplitterLayout(d.theme)
	d.layout.SetOrientation(gxui.Vertical)
	d.layout.AddChild(d.dirs)
	d.layout.AddChild(d.files)

	d.dirs.OnSelectionChanged(func(selection gxui.AdapterItem) {
		d.filesAdapter = fileList(selection.(string))
		d.files.SetAdapter(d.filesAdapter)
	})
}

func (d *ProjectTree) Button() gxui.Button {
	return d.button
}

func (d *ProjectTree) SetRoot(path string) {
	d.dirsAdapter = dirTree(path)
	d.dirs.SetAdapter(d.dirsAdapter)
	d.filesAdapter = fileList(path)
	d.files.SetAdapter(d.filesAdapter)
}

func (d *ProjectTree) SetProject(project string) {
	for _, proj := range settings.Projects() {
		if proj.Name == project {
			d.SetRoot(proj.Path)
			return
		}
	}
	panic(fmt.Errorf("Could not find project %s", project))
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

type splitterLayout struct {
	mixins.SplitterLayout
	theme *basic.Theme
}

func newSplitterLayout(theme *basic.Theme) *splitterLayout {
	layout := &splitterLayout{
		theme: theme,
	}
	layout.Init(layout, theme)
	return layout
}

func (l *splitterLayout) CreateSplitterBar() gxui.Control {
	b := &mixins.SplitterBar{}
	b.Init(b, l.theme)
	b.SetBackgroundColor(splitterBarBackgroundColor)
	b.SetForegroundColor(splitterBarForegroundColor)
	return b
}

func (l *splitterLayout) DesiredSize(min, max math.Size) math.Size {
	width := 30 * l.theme.DefaultMonospaceFont().GlyphMaxSize().W
	if min.W > width {
		width = min.W
	}
	if max.W < width {
		width = max.W
	}
	return math.Size{
		W: width,
		H: max.H,
	}
}

type fsAdapter struct {
	gxui.AdapterBase
	path     string
	children []string
	click    func(event gxui.MouseEvent)
}

func fsList(dirPath string, dirs bool) fsAdapter {
	dir := fsAdapter{
		path: dirPath,
	}
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && path != dirPath {
			use := (dirs && info.IsDir()) || (!dirs && !info.IsDir())
			if use && !strings.HasPrefix(info.Name(), ".") {
				dir.children = append(dir.children, path)
			}
			if info.IsDir() {
				return filepath.SkipDir
			}
		}
		return nil
	})
	return dir
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

type dirTreeAdapter struct {
	fsAdapter
}

func dirTree(dirPath string) *dirTreeAdapter {
	return &dirTreeAdapter{
		fsAdapter: fsList(dirPath, true),
	}
}

func (a *dirTreeAdapter) Item() gxui.AdapterItem {
	return a.path
}

func (a *dirTreeAdapter) NodeAt(index int) gxui.TreeNode {
	return dirTree(a.children[index])
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
