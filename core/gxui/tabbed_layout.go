// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/ui"
)

// TabbedLayout wraps up gxui linear layouts to match ui.Layout
type TabbedLayout struct {
	holder gxui.PanelHolder
}

func (l TabbedLayout) Control() gxui.Control {
	return l.holder
}

func (l TabbedLayout) child(el interface{}) (gxui.Control, error) {
	c, err := control(el)
	if err != nil {
		return nil, err
	}
	if l.holder.PanelIndex(c) == -1 {
		return nil, fmt.Errorf("gxui: %v not found as a child of %v", el, l)
	}

	return c, nil
}

func (l TabbedLayout) SetFocus(el interface{}) error {
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

func (l TabbedLayout) addIndex(cfg ui.LayoutCfg) int {
	if cfg.NextTo == nil {
		return 0
	}
	sibling, err := l.child(cfg.NextTo.Sibling)
	if err == nil {
		idx := l.holder.PanelIndex(sibling)
		if cfg.NextTo.Side == ui.Right {
			return idx + 1
		}
		return idx
	}
	if cfg.NextTo.Side == ui.Right {
		return l.holder.PanelCount()
	}
	return 0
}

func (l TabbedLayout) Add(child interface{}, opts ...ui.LayoutOpt) error {
	ctrl, err := control(child)
	if err != nil {
		return err
	}
	var cfg ui.LayoutCfg
	for _, o := range opts {
		cfg = o(cfg)
	}
	if cfg.Name == nil {
		return fmt.Errorf("gxui: cannot create tab without a name (see ui.LayoutName docs)")
	}
	l.holder.AddPanelAt(ctrl, *cfg.Name, l.addIndex(cfg))
	return nil
}

func (l TabbedLayout) Remove(child interface{}) error {
	c, err := l.child(child)
	if err != nil {
		return err
	}
	l.holder.RemovePanel(c)
	return nil
}
