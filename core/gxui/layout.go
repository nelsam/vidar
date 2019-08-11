// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins/base"
	"github.com/nelsam/gxui/mixins/parts"
	"github.com/nelsam/vidar/ui"
)

// LinearLayout wraps up gxui linear layouts to match ui.Layout
type LinearLayout struct {
	layout gxui.LinearLayout
}

func (l LinearLayout) Control() gxui.Control {
	return l.layout
}

func (l LinearLayout) child(el interface{}) (gxui.Control, error) {
	c, err := control(el)
	if err != nil {
		return nil, err
	}
	if l.layout.Children().Find(c) == nil {
		return nil, fmt.Errorf("gxui: %v not found as a child of %v", el, l)
	}

	return c, nil
}

func (l LinearLayout) SetFocus(el interface{}) error {
	c, err := l.child(el)
	if err != nil {
		return err
	}
	f, ok := c.(gxui.Focusable)
	if !ok {
		return fmt.Errorf("gxui: cannot focus %v (control type %T) because it is not gxui.Focusable", el, c)
	}
	gxui.SetFocus(f)
	return nil
}

func (l LinearLayout) Add(child interface{}, opts ...ui.LayoutOpt) error {
	ctrl, err := control(child)
	if err != nil {
		return err
	}
	var cfg ui.LayoutCfg
	for _, o := range opts {
		cfg = o(cfg)
	}
	c := l.layout.AddChild(ctrl)
	if cfg.Offset != nil {
		c.Offset = math.Point(*cfg.Offset)
	}
	return nil
}

func (l LinearLayout) Remove(child interface{}) error {
	ctrl, err := l.child(child)
	if err != nil {
		return err
	}
	l.layout.RemoveChild(ctrl)
	return nil
}

// ManualLayout is a custom gxui layout for the creator.ManualLayout method.
type ManualLayout struct {
	base.Container
	parts.BackgroundBorderPainter
}

// newManualLayout creates a ManualLayout.
func newManualLayout(creator Creator) (*ManualLayout, error) {
	layout := &ManualLayout{}

	_, theme := creator.Raw()
	layout.Container.Init(layout, theme)
	layout.BackgroundBorderPainter.Init(layout)
	layout.SetMouseEventTarget(true)
	layout.SetBackgroundBrush(gxui.TransparentBrush)
	layout.SetBorderPen(gxui.TransparentPen)

	return layout, nil
}

func (l *ManualLayout) Control() gxui.Control {
	return l
}

func (l *ManualLayout) child(el interface{}) (gxui.Control, error) {
	c, err := control(el)
	if err != nil {
		return nil, err
	}
	if l.Children().Find(c) == nil {
		return nil, fmt.Errorf("gxui: %v not found as a child of %v", el, l)
	}

	return c, nil
}

func (l *ManualLayout) SetFocus(el interface{}) error {
	c, err := l.child(el)
	if err != nil {
		return err
	}
	f, ok := c.(gxui.Focusable)
	if !ok {
		return fmt.Errorf("gxui: cannot focus %v (control type %T) because it is not gxui.Focusable", el, c)
	}
	gxui.SetFocus(f)
	return nil
}

func (l *ManualLayout) Add(child interface{}, opts ...ui.LayoutOpt) error {
	ctrl, err := control(child)
	if err != nil {
		return err
	}
	c := l.Container.AddChild(ctrl)

	var cfg ui.LayoutCfg
	for _, o := range opts {
		cfg = o(cfg)
	}
	if cfg.Offset != nil {
		c.Offset = math.Point(*cfg.Offset)
	}
	return nil
}

func (l *ManualLayout) Remove(child interface{}) error {
	c, err := l.child(child)
	if err != nil {
		return err
	}
	l.Container.RemoveChild(c)
	return nil
}

func (c *ManualLayout) DesiredSize(_, max math.Size) math.Size {
	return max
}

func (c *ManualLayout) LayoutChildren() {
	maxSize := c.Size().Contract(c.Padding())
	for _, child := range c.Children() {
		margin := child.Control.Margin()
		desiredSize := child.Control.DesiredSize(math.ZeroSize, maxSize.Contract(margin))
		child.Control.SetSize(desiredSize)
	}
}

func (c *ManualLayout) Paint(canvas gxui.Canvas) {
	rect := c.Size().Rect()
	c.BackgroundBorderPainter.PaintBackground(canvas, rect)
	c.PaintChildren.Paint(canvas)
	c.BackgroundBorderPainter.PaintBorder(canvas, rect)
}
