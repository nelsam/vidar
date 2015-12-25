// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package controller

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/controller/editor"
	"github.com/nelsam/vidar/controller/navigator"
)

type Navigator interface {
	gxui.Control
	Add(navigator.Pane)
}

type Editor interface {
	gxui.Control
	Focus()
	CurrentFile() string
	Open(path string, cursor int)
}

type Elementer interface {
	Elements() []interface{}
}

type Controller struct {
	mixins.LinearLayout

	theme     *basic.Theme
	driver    gxui.Driver
	font      gxui.Font
	navigator Navigator
	editor    Editor
}

func New(driver gxui.Driver, theme *basic.Theme, font gxui.Font) *Controller {
	controller := new(Controller)
	controller.Init(driver, theme, font)
	return controller
}

func (c *Controller) Init(driver gxui.Driver, theme *basic.Theme, font gxui.Font) {
	c.LinearLayout.Init(c, theme)
	c.driver = driver
	c.theme = theme
	c.font = font

	c.SetDirection(gxui.LeftToRight)
	c.navigator = navigator.New(driver, theme, c.Execute)
	c.AddChild(c.navigator)
	c.editor = editor.New(driver, theme, font)
	c.AddChild(c.editor)
}

// Navigator returns c's Navigator instance.
func (c *Controller) Navigator() Navigator {
	return c.navigator
}

// Editor returns c's Editor instance
func (c *Controller) Editor() Editor {
	return c.editor
}

// Execute implements "../commander".Controller.
func (c *Controller) Execute(command commands.Command) {
	executed, _ := execRecursively(command, c)
	if !executed {
		panic(fmt.Errorf("Command %s ran without executing", command.Name()))
	}
	c.Editor().Focus()
}

func execRecursively(command commands.Command, element interface{}) (executed, consume bool) {
	executed, consume = command.Exec(element)
	if consume {
		return
	}
	var childExecuted bool
	if parent, ok := element.(gxui.Parent); ok {
		for _, child := range parent.Children() {
			childExecuted, consume = execRecursively(command, child.Control)
			executed = executed || childExecuted
			if consume {
				return
			}
		}
	}
	if elementer, ok := element.(Elementer); ok {
		for _, element := range elementer.Elements() {
			childExecuted, consume = execRecursively(command, element)
			executed = executed || childExecuted
			if consume {
				return
			}
		}
	}
	return
}
