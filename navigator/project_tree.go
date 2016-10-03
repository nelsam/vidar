// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

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

	dirs *dirTree

	layout gxui.LinearLayout
}

func NewProjectTree(driver gxui.Driver, theme *basic.Theme) *ProjectTree {
	tree := &ProjectTree{
		driver: driver,
		theme:  theme,
		button: createIconButton(driver, theme, "folder.png"),
		layout: theme.CreateLinearLayout(),
	}
	tree.layout.SetDirection(gxui.TopToBottom)

	tree.SetProject(settings.DefaultProject)

	return tree
}

func (p *ProjectTree) Button() gxui.Button {
	return p.button
}

func (p *ProjectTree) SetRoot(path string) {
	p.layout.RemoveAll()
	p.dirs = newDirTree(p.theme, path)
	p.dirs.Load()
	p.layout.AddChild(p.dirs)
}

func (p *ProjectTree) SetProject(project settings.Project) {
	p.SetRoot(project.Path)
}

func (p *ProjectTree) Open(filePath string) {
	//dir, _ := filepath.Split(filePath)
}

func (p *ProjectTree) Frame() gxui.Control {
	return p.layout
}

func (p *ProjectTree) OnComplete(onComplete func(controller.Executor)) {
	//cmd := commands.NewFileOpener(p.driver, p.theme)
}

type dirTree struct {
	mixins.LinearLayout

	theme gxui.Theme
	path  string
}

func newDirTree(theme gxui.Theme, path string) *dirTree {
	t := &dirTree{
		theme: theme,
		path:  path,
	}
	t.Init(t, theme)
	t.SetDirection(gxui.TopToBottom)
	return t
}

func (d *dirTree) DesiredSize(min, max math.Size) math.Size {
	s := d.LinearLayout.DesiredSize(min, max)
	width := 20 * d.theme.DefaultMonospaceFont().GlyphMaxSize().W
	if min.W > width {
		width = min.W
	}
	if max.W < width {
		width = max.W
	}
	s.W = width
	return s
}

func (d *dirTree) Load() error {
	d.RemoveAll()
	finfos, err := ioutil.ReadDir(d.path)
	if err != nil {
		return err
	}
	for _, finfo := range finfos {
		if !finfo.IsDir() {
			continue
		}
		if strings.HasPrefix(finfo.Name(), ".") {
			continue
		}
		layout := d.theme.CreateLinearLayout()
		layout.SetDirection(gxui.TopToBottom)
		d.AddChild(layout)
		button := newDirButton(d.theme.(*basic.Theme), finfo.Name())
		path := filepath.Join(d.path, button.Text())

		// TODO: move this logic.  It's here temporarily while I spike out
		// alternative tree logic.
		finfos, err := ioutil.ReadDir(path)
		if err == nil {
			for _, finfo := range finfos {
				if finfo.IsDir() {
					button.SetText(button.Text() + " ►")
					break
				}
			}
		}
		layout.AddChild(button)
		subTree := newDirTree(d.theme, path)
		subTree.SetMargin(math.Spacing{L: 10})
		button.OnClick(func(gxui.MouseEvent) {
			if len(layout.Children()) > 1 {
				button.SetText(strings.TrimSuffix(button.Text(), " ▼") + " ►")
				layout.RemoveChild(subTree)
				return
			}
			subTree.Load()
			button.SetText(strings.TrimSuffix(button.Text(), " ►") + " ▼")
			layout.AddChild(subTree)
		})
	}
	return nil
}

type dirButton struct {
	mixins.Button

	theme *basic.Theme
}

func newDirButton(theme *basic.Theme, name string) *dirButton {
	d := &dirButton{
		theme: theme,
	}
	d.Init(d, theme)
	d.SetText(name)
	d.Label().SetColor(dirColor)
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

func (d *dirButton) DesiredSize(min, max math.Size) math.Size {
	s := d.Button.DesiredSize(min, max)
	s.W = max.W
	return s
}

func (d *dirButton) Style() (s basic.Style) {
	defer func() {
		s.FontColor = d.Label().Color()
	}()
	if d.IsMouseDown(gxui.MouseButtonLeft) && d.IsMouseOver() {
		return d.theme.ButtonPressedStyle
	}
	if d.IsMouseOver() {
		return d.theme.ButtonOverStyle
	}
	return d.theme.ButtonDefaultStyle
}

func (d *dirButton) Paint(canvas gxui.Canvas) {
	style := d.Style()
	if l := d.Label(); l != nil {
		l.SetColor(style.FontColor)
	}

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
	//canvas.DrawLines(poly, style.Pen)
}
