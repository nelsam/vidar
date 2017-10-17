// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package controller

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/editor"
)

type Navigator interface {
	gxui.Control
}

type Editor interface {
	gxui.Control
	CurrentFile() string
	CurrentEditor() *editor.CodeEditor
	Open(path string, offset int) (editor *editor.CodeEditor, existed bool)
}

type Controller struct {
	mixins.LinearLayout

	theme     *basic.Theme
	driver    gxui.Driver
	font      gxui.Font
	navigator Navigator
	editor    Editor
}

func New(driver gxui.Driver, theme *basic.Theme) *Controller {
	controller := new(Controller)
	controller.Init(driver, theme)
	return controller
}

func (c *Controller) Init(driver gxui.Driver, theme *basic.Theme) {
	c.LinearLayout.Init(c, theme)
	c.driver = driver
	c.theme = theme

	c.SetDirection(gxui.LeftToRight)
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
func (c *Controller) Editor() Editor {
	return c.editor
}

// SetEditor sets c's Editor instance.
func (c *Controller) SetEditor(editor Editor) {
	if c.editor != nil {
		c.RemoveChild(c.editor)
	}
	c.editor = editor
	if c.editor != nil {
		c.AddChild(c.editor)
	}
}
