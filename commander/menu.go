// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/mixins/parts"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
)

type Boundser interface {
	Bounds() math.Rect
}

type menuBar struct {
	mixins.LinearLayout

	commander *Commander
	theme     *basic.Theme
	menus     map[string]*menu
}

func newMenuBar(commander *Commander, theme *basic.Theme) *menuBar {
	m := &menuBar{
		commander: commander,
		theme:     theme,
		menus:     make(map[string]*menu),
	}
	m.Init(m, theme)
	m.SetDirection(gxui.LeftToRight)
	m.SetBorderPen(theme.ButtonDefaultStyle.Pen)
	return m
}

func (m *menuBar) Add(command bind.Command, bindings ...gxui.KeyboardEvent) {
	menu, ok := m.menus[command.Menu()]
	if !ok {
		menu = newMenu(m.commander, m.theme)
		m.menus[command.Menu()] = menu
		button := newMenuButton(m.commander, m.theme, command.Menu())
		child := m.AddChild(button)
		button.SetMenu(child, menu)
	}
	menu.Add(command, bindings...)
}

func (m *menuBar) Clear() {
	m.RemoveAll()
	m.menus = make(map[string]*menu)
}

func (m *menuBar) Paint(canvas gxui.Canvas) {
	rect := m.Size().Rect()
	m.BackgroundBorderPainter.PaintBackground(canvas, rect)
	m.PaintChildren.Paint(canvas)
	painterRect := math.Rect{
		Min: math.Point{
			X: rect.Min.X,
			Y: rect.Max.Y - 1,
		},
		Max: rect.Max,
	}
	m.BackgroundBorderPainter.PaintBorder(canvas, painterRect)
}

func (m *menuBar) DesiredSize(min, max math.Size) math.Size {
	size := m.LinearLayout.DesiredSize(min, max)
	size.W = max.W
	return size
}

type menu struct {
	parts.Focusable
	mixins.LinearLayout

	commander *Commander
	theme     *basic.Theme
}

func newMenu(commander *Commander, theme *basic.Theme) *menu {
	m := &menu{
		commander: commander,
		theme:     theme,
	}
	m.Focusable.Init(m)
	m.LinearLayout.Init(m, theme)
	m.SetBackgroundBrush(theme.ButtonDefaultStyle.Brush)
	m.SetBorderPen(theme.ButtonDefaultStyle.Pen)
	return m
}

func (m *menu) Add(command bind.Command, bindings ...gxui.KeyboardEvent) {
	item := newMenuItem(m.theme, command.Name(), bindings...)
	m.AddChild(item)
	item.OnClick(func(gxui.MouseEvent) {
		if m.commander.box.Run(command) {
			gxui.SetFocus(m.commander.box.input)
			return
		}
		if executor, ok := command.(bind.Executor); ok {
			m.commander.Execute(executor)
		}
		m.commander.box.Finish()
	})
}

type menuItem struct {
	mixins.Button

	theme *basic.Theme
}

func newMenuItem(theme *basic.Theme, name string, bindings ...gxui.KeyboardEvent) *menuItem {
	b := &menuItem{
		theme: theme,
	}
	b.Init(b, theme)
	for i, binding := range bindings {
		if i == 0 {
			name += "   "
		} else {
			name += ", "
		}
		parts := make([]string, 0, 5)
		if binding.Modifier.Control() {
			parts = append(parts, "Ctrl")
		}
		if binding.Modifier.Super() {
			parts = append(parts, "Cmd")
		}
		if binding.Modifier.Alt() {
			parts = append(parts, "Alt")
		}
		if binding.Modifier.Shift() {
			parts = append(parts, "Shift")
		}
		parts = append(parts, binding.Key.String())
		name += strings.Join(parts, "-")
	}
	b.SetText(name)
	b.SetPadding(math.Spacing{L: 1, R: 1, B: 1, T: 1})
	b.SetMargin(math.Spacing{L: 2, R: 2})
	b.OnMouseEnter(func(gxui.MouseEvent) { b.Redraw() })
	b.OnMouseExit(func(gxui.MouseEvent) { b.Redraw() })
	b.OnMouseDown(func(gxui.MouseEvent) { b.Redraw() })
	b.OnMouseUp(func(gxui.MouseEvent) { b.Redraw() })
	b.OnGainedFocus(b.Redraw)
	b.OnLostFocus(b.Redraw)
	return b
}

func (b *menuItem) Style() basic.Style {
	if b.IsMouseDown(gxui.MouseButtonLeft) && b.IsMouseOver() {
		return b.theme.ButtonPressedStyle
	}
	if b.IsMouseOver() {
		return b.theme.ButtonOverStyle
	}
	return b.theme.ButtonDefaultStyle
}

func (b *menuItem) Paint(canvas gxui.Canvas) {
	style := b.Style()
	if l := b.Label(); l != nil {
		l.SetColor(style.FontColor)
	}

	rect := b.Size().Rect()
	b.BackgroundBorderPainter.PaintBackground(canvas, rect)
	b.PaintChildren.Paint(canvas)
}

type menuButton struct {
	mixins.Button

	menuParent gxui.Container
	theme      *basic.Theme
}

func newMenuButton(menuParent gxui.Container, theme *basic.Theme, name string) *menuButton {
	b := &menuButton{
		menuParent: menuParent,
		theme:      theme,
	}
	b.Init(b, theme)
	b.SetText(name)
	b.SetPadding(math.Spacing{L: 1, R: 1, B: 1, T: 1})
	b.SetMargin(math.Spacing{L: 3})
	b.SetBackgroundBrush(b.theme.ButtonDefaultStyle.Brush)
	b.SetBorderPen(b.theme.ButtonDefaultStyle.Pen)
	b.OnMouseEnter(func(gxui.MouseEvent) { b.Redraw() })
	b.OnMouseExit(func(gxui.MouseEvent) { b.Redraw() })
	b.OnMouseDown(func(gxui.MouseEvent) { b.Redraw() })
	b.OnMouseUp(func(gxui.MouseEvent) { b.Redraw() })
	b.OnGainedFocus(b.Redraw)
	b.OnLostFocus(b.Redraw)
	return b
}

func (b *menuButton) SetMenu(boundser Boundser, menu *menu) {
	b.OnClick(func(gxui.MouseEvent) {
		if b.menuParent.Children().IndexOf(menu) >= 0 {
			b.menuParent.RemoveChild(menu)
			return
		}
		bounds := boundser.Bounds()
		child := b.menuParent.AddChild(menu)
		offset := math.Point{
			X: bounds.Min.X,
			Y: bounds.Max.Y,
		}
		child.Offset = offset
		gxui.SetFocus(menu)
	})
	menu.OnLostFocus(func() {
		if b.menuParent.Children().IndexOf(menu) >= 0 {
			b.menuParent.RemoveChild(menu)
		}
	})
}

func (b *menuButton) Style() basic.Style {
	if b.IsMouseDown(gxui.MouseButtonLeft) && b.IsMouseOver() {
		return b.theme.ButtonPressedStyle
	}
	if b.IsMouseOver() {
		return b.theme.ButtonOverStyle
	}
	return b.theme.ButtonDefaultStyle
}

func (b *menuButton) Paint(canvas gxui.Canvas) {
	style := b.Style()
	if l := b.Label(); l != nil {
		l.SetColor(style.FontColor)
	}

	rect := b.Size().Rect()
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
	b.PaintChildren.Paint(canvas)
	canvas.DrawLines(poly, style.Pen)
}
