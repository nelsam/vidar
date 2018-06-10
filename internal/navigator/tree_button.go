// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
)

var preferred = []dropdownCharSet{
	{
		expanded:  '▼',
		collapsed: '►',
	},
	{
		expanded:  '∨',
		collapsed: '›',
	},
	{
		expanded:  '◊',
		collapsed: '›',
	},
	{
		expanded:  'v',
		collapsed: '>',
	},
}

type dropdownCharSet struct {
	expanded, collapsed rune
}

type treeButton struct {
	mixins.Button

	driver gxui.Driver
	theme  *basic.Theme
	drop   *mixins.Label

	dropSet dropdownCharSet
}

type indexable interface {
	Index(r rune) int
}

func newTreeButton(driver gxui.Driver, theme *basic.Theme, name string) *treeButton {
	d := &treeButton{
		driver: driver,
		theme:  theme,
		drop:   &mixins.Label{},
	}
	d.drop.Init(d.drop, d.theme, d.theme.DefaultMonospaceFont(), dropColor)
	d.Init(d, theme)

	d.chooseDropSet()

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

func (d *treeButton) chooseDropSet() {
	font := d.theme.DefaultFont()
	for _, d.dropSet = range preferred {
		if font.Index(d.dropSet.collapsed) == 0 {
			continue
		}
		if font.Index(d.dropSet.expanded) == 0 {
			continue
		}
		return
	}
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
		text = fmt.Sprintf(" %c", d.dropSet.collapsed)
	}
	d.drop.SetText(text)
}

func (d *treeButton) Expanded() bool {
	return d.Expandable() && d.drop.Text() == fmt.Sprintf(" %c", d.dropSet.expanded)
}

func (d *treeButton) Expandable() bool {
	return d.drop.Text() != ""
}

func (d *treeButton) Expand() {
	d.drop.SetText(fmt.Sprintf(" %c", d.dropSet.expanded))
}

func (d *treeButton) Collapse() {
	d.drop.SetText(fmt.Sprintf(" %c", d.dropSet.collapsed))
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
