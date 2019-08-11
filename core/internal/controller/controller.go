// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package controller

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/ui"
)

type deprecatedGXUICoupling interface {
	Control() gxui.Control
}

type Navigator interface {
	gxui.Control
}

type MultiEditor interface {
	gxui.Control
	CurrentFile() string
	CurrentEditor() input.Editor
	Open(path string) (editor input.Editor, existed bool)
}

type Creator interface {
	LinearLayout(ui.Direction) (ui.Layout, error)
}

type Controller struct {
	mixins.LinearLayout

	navigator Navigator
	editor    MultiEditor
}

func New(creator Creator) (*Controller, error) {
	layout, err := creator.LinearLayout(ui.Left)
	if err != nil {
		return nil, err
	}
	ll := layout.(deprecatedGXUICoupling).Control().(*mixins.LinearLayout)
	c := &Controller{LinearLayout: *ll}
	return c, nil
}

// Navigator returns c's Navigator instance.
func (c *Controller) Navigator() Navigator {
	return c.navigator
}

func (c *Controller) Elements() []interface{} {
	return []interface{}{
		c.navigator,
		c.editor,
	}
}

// SetNavigator sets c's Navigator instance.
func (c *Controller) SetNavigator(navigator Navigator) {
	if c.navigator != nil {
		c.RemoveChild(c.navigator)
	}
	c.navigator = navigator
	if c.navigator != nil {
		c.AddChild(c.navigator)

		// Reset the editor, since it should come after the navigator.
		c.SetEditor(c.editor)
	}
}

// Editor returns c's Editor instance.
func (c *Controller) Editor() MultiEditor {
	return c.editor
}

// SetEditor sets c's Editor instance.
func (c *Controller) SetEditor(editor MultiEditor) {
	if c.editor != nil {
		c.RemoveChild(c.editor)
	}
	c.editor = editor
	if c.editor != nil {
		c.AddChild(c.editor)
	}
}
