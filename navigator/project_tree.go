// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
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
	dropColor = gxui.Color{
		R: 0.7,
		G: 0.7,
		B: 0.1,
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

	callback func(controller.Executor)
	dirs     *directory
	toc      gxui.Control

	layout *splitterLayout
}

func NewProjectTree(driver gxui.Driver, theme *basic.Theme) *ProjectTree {
	tree := &ProjectTree{
		driver: driver,
		theme:  theme,
		button: createIconButton(driver, theme, "folder.png"),
		layout: newSplitterLayout(theme),
	}
	tree.layout.SetOrientation(gxui.Vertical)

	tree.SetProject(settings.DefaultProject)

	return tree
}

func (p *ProjectTree) Button() gxui.Button {
	return p.button
}

func (p *ProjectTree) SetRoot(path string) {
	p.layout.RemoveAll()
	p.toc = nil

	p.dirs = newDirectory(p, path)
	scrollable := p.theme.CreateScrollLayout()
	// Disable horiz scrolling until we can figure out an accurate
	// way to calculate our width.
	scrollable.SetScrollAxis(false, true)
	scrollable.SetChild(p.dirs)
	p.layout.AddChild(scrollable)
	p.layout.SetChildWeight(p.dirs, 1)

	// Expand the top level
	p.dirs.button.Click(gxui.MouseEvent{})
}

func (p *ProjectTree) SetProject(project settings.Project) {
	p.SetRoot(project.Path)
}

func (p *ProjectTree) Open(filePath string) {
	dir, _ := filepath.Split(filePath)
	p.dirs.ExpandTo(dir)
}

func (p *ProjectTree) Frame() gxui.Control {
	return p.layout
}

func (p *ProjectTree) OnComplete(callback func(controller.Executor)) {
	p.callback = callback
	go attachCallback(p.toc, callback)
}

type parent interface {
	Children() gxui.Children
}

type irrespParent interface {
	MissingChild() gxui.Control
}

type selectionButton interface {
	OnSelected(func(controller.Executor))
}

func attachCallback(control gxui.Control, callback func(controller.Executor)) {
	if b, ok := control.(selectionButton); ok {
		b.OnSelected(callback)
	}
	if p, ok := control.(parent); ok {
		for _, c := range p.Children() {
			attachCallback(c.Control, callback)
		}
	}
	if p, ok := control.(irrespParent); ok {
		attachCallback(p.MissingChild(), callback)
	}
}

type directory struct {
	mixins.LinearLayout

	driver gxui.Driver
	button *treeButton
	tree   *dirTree

	// length is an atomically updated list of child nodes of
	// this directory.  Only access via atomics.
	length int64
}

func newDirectory(projTree *ProjectTree, path string) *directory {
	driver := projTree.driver
	theme := projTree.theme

	button := newTreeButton(driver, theme, filepath.Base(path))
	tree := newDirTree(projTree, path)
	tree.SetMargin(math.Spacing{L: 10})
	d := &directory{
		driver: driver,
		button: button,
		tree:   tree,
	}
	d.Init(d, theme)
	d.AddChild(button)
	button.OnClick(func(gxui.MouseEvent) {
		if projTree.toc != nil {
			projTree.layout.RemoveChild(projTree.toc)
		}
		toc := NewTOC(projTree.driver, projTree.theme, path)
		if projTree.callback != nil {
			go attachCallback(toc, projTree.callback)
		}
		scrollable := theme.CreateScrollLayout()
		// Disable horiz scrolling until we can figure out an accurate
		// way to calculate our width.
		scrollable.SetScrollAxis(false, true)
		scrollable.SetChild(toc)
		projTree.toc = scrollable
		projTree.layout.AddChild(projTree.toc)
		projTree.layout.SetChildWeight(projTree.toc, 2)
		if d.Length() == 0 {
			return
		}
		if d.tree.Attached() {
			d.button.Collapse()
			d.RemoveChild(d.tree)
			return
		}
		d.tree.Load()
		d.button.Expand()
		d.AddChild(d.tree)
	})
	d.reload()
	return d
}

//func (d *directory) Child(dir string)

func (d *directory) ExpandTo(dir string) {
	if !strings.HasPrefix(dir, d.tree.path) {
		return
	}
	d.button.Click(gxui.MouseEvent{})
	for _, child := range d.tree.Dirs() {
		child.ExpandTo(dir)
	}
}

func (d *directory) Length() int64 {
	return atomic.LoadInt64(&d.length)
}

func (d *directory) updateExpandable(children int64) {
	if children == 0 {
		d.button.SetExpandable(false)
		return
	}
	d.button.SetExpandable(true)
}

func (d *directory) reload() {
	finfos, err := ioutil.ReadDir(d.tree.path)
	if err != nil {
		log.Printf("Unexpected error reading directory %s: %s", d.tree.path, err)
		return
	}
	defer d.driver.Call(d.Redraw)

	children := int64(0)
	for _, finfo := range finfos {
		if finfo.IsDir() {
			children++
		}
	}

	d.updateExpandable(children)
	atomic.StoreInt64(&d.length, children)
	if d.tree.Attached() {
		d.tree.parse(finfos)
	}
}

type dirTree struct {
	mixins.LinearLayout

	projTree *ProjectTree
	driver   gxui.Driver
	theme    gxui.Theme
	path     string
}

func newDirTree(projTree *ProjectTree, path string) *dirTree {
	t := &dirTree{
		projTree: projTree,
		driver:   projTree.driver,
		theme:    projTree.theme,
		path:     path,
	}
	t.Init(t, projTree.theme)
	t.SetDirection(gxui.TopToBottom)
	return t
}

func (d *dirTree) Dirs() (dirs []*directory) {
	for _, c := range d.Children() {
		dirs = append(dirs, c.Control.(*directory))
	}
	return dirs
}

func (d *dirTree) Load() error {
	finfos, err := ioutil.ReadDir(d.path)
	if err != nil {
		return err
	}
	d.parse(finfos)
	return nil
}

func (d *dirTree) parse(finfos []os.FileInfo) {
	d.RemoveAll()
	for _, finfo := range finfos {
		if !finfo.IsDir() {
			continue
		}
		if strings.HasPrefix(finfo.Name(), ".") {
			continue
		}
		dir := newDirectory(d.projTree, filepath.Join(d.path, finfo.Name()))
		d.AddChild(dir)
	}
}

type treeButton struct {
	mixins.Button

	driver gxui.Driver
	theme  *basic.Theme
	drop   *mixins.Label
}

func newTreeButton(driver gxui.Driver, theme *basic.Theme, name string) *treeButton {
	d := &treeButton{
		driver: driver,
		theme:  theme,
		drop:   &mixins.Label{},
	}
	d.drop.Init(d.drop, d.theme, d.theme.DefaultMonospaceFont(), dropColor)
	d.Init(d, theme)
	d.SetDirection(gxui.LeftToRight)
	d.SetText(name)
	d.Label().SetColor(dirColor)
	d.AddChild(d.drop)
	d.SetPadding(math.Spacing{L: 1, R: 1, B: 1, T: 1})
	d.SetMargin(math.Spacing{L: 3})
	d.SetBackgroundBrush(d.theme.ButtonDefaultStyle.Brush)
	d.OnMouseEnter(func(gxui.MouseEvent) { d.Redraw() })
	d.OnMouseExit(func(gxui.MouseEvent) { d.Redraw() })
	d.OnMouseDown(func(gxui.MouseEvent) { d.Redraw() })
	d.OnMouseUp(func(gxui.MouseEvent) { d.Redraw() })
	d.OnGainedFocus(d.Redraw)
	d.OnLostFocus(d.Redraw)
	return d
}

func (d *treeButton) SetExpandable(expandable bool) {
	if expandable && d.drop.Text() != "" {
		return
	}
	if !expandable && d.drop.Text() == "" {
		return
	}
	text := ""
	if expandable {
		text = " ►"
	}
	d.drop.SetText(text)
}

func (d *treeButton) Expandable() bool {
	return d.drop.Text() != ""
}

func (d *treeButton) Expand() {
	d.drop.SetText(" ▼")
}

func (d *treeButton) Collapse() {
	d.drop.SetText(" ►")
}

func (d *treeButton) DesiredSize(min, max math.Size) math.Size {
	s := d.Button.DesiredSize(min, max)
	s.W = max.W
	return s
}

func (d *treeButton) Style() (s basic.Style) {
	if d.IsMouseDown(gxui.MouseButtonLeft) && d.IsMouseOver() {
		return d.theme.ButtonPressedStyle
	}
	if d.IsMouseOver() {
		return d.theme.ButtonOverStyle
	}
	return d.theme.ButtonDefaultStyle
}

func (d *treeButton) Paint(canvas gxui.Canvas) {
	style := d.Style()

	rect := d.Size().Rect()
	poly := gxui.Polygon{
		{Position: math.Point{
			X: rect.Min.X,
			Y: rect.Max.Y,
		}},
		{Position: math.Point{
			X: rect.Min.X,
			Y: rect.Min.Y,
		}},
		{Position: math.Point{
			X: rect.Max.X,
			Y: rect.Min.Y,
		}},
		{Position: math.Point{
			X: rect.Max.X,
			Y: rect.Max.Y,
		}},
	}
	canvas.DrawPolygon(poly, gxui.TransparentPen, style.Brush)
	d.PaintChildren.Paint(canvas)
}

type splitterLayout struct {
	mixins.SplitterLayout

	theme gxui.Theme
}

func newSplitterLayout(theme gxui.Theme) *splitterLayout {
	l := &splitterLayout{
		theme: theme,
	}
	l.Init(l, theme)
	return l
}

func (l *splitterLayout) DesiredSize(min, max math.Size) math.Size {
	s := l.SplitterLayout.DesiredSize(min, max)
	width := 20 * l.theme.DefaultMonospaceFont().GlyphMaxSize().W
	if min.W > width {
		width = min.W
	}
	if max.W < width {
		width = max.W
	}
	s.W = width
	return s
}

func (l *splitterLayout) CreateSplitterBar() gxui.Control {
	bar := l.SplitterLayout.CreateSplitterBar()
	return bar
}
