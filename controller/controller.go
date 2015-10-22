// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package controller

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui_playground/controller/editor"
	"github.com/nelsam/gxui_playground/controller/navigator"
)

// Command is a command that executes against the Controller or one of
// its children.
type Command interface {
	// Start starts the command.  The returned gxui.Control can be
	// used to display the status of the command.  The elements
	// returned by Next() should have OnUnfocus() triggers to update
	// this element.
	//
	// If no display element is necessary, or if display will be taken
	// care of by the input elements, Start should return nil.
	Start(gxui.Control) gxui.Control

	// Name returns the name of the command
	Name() string

	// Next returns the next control element for reading user input.
	// By default, this is called every time the commander receives a
	// gxui.KeyEnter event in KeyPress.  If there are situations where
	// this is not the desired behavior, the gxui.Control should
	// consume the event.  If more keys are required to signal the end
	// of input, the control should implement Completer.
	//
	// When no more user input is required, Next should return nil.
	Next() gxui.Focusable

	// Exec executes the command on the provided element.  It should
	// return true when it no longer needs to descend into nested
	// elements.  For example, if the command needs to execute against
	// a gxui.Editor, it should return false until it is passed a
	// gxui.Editor, then perform its action and return true.
	Exec(interface{}) (consume bool)
}

type Navigator interface {
	gxui.Control
	Add(navigator.Pane)
}

type Editor interface {
	gxui.PanelHolder
	Focus()
	CurrentFile() string
	Open(file string)
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
	c.navigator = navigator.New(driver, theme, font)
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

// Execute implements
// "github.com/nelsam/gxui_playground/commander".Controller.
func (c *Controller) Execute(command Command) {
	execRecursively(command, c)
	c.Editor().Focus()
}

func execRecursively(command Command, element interface{}) (consume bool) {
	consume = command.Exec(element)
	if parent, ok := element.(gxui.Parent); ok && !consume {
		for _, child := range parent.Children() {
			consume = execRecursively(command, child.Control)
			if consume {
				return
			}
		}
	}
	if elementer, ok := element.(Elementer); ok {
		for _, element := range elementer.Elements() {
			consume = execRecursively(command, element)
			if consume {
				return
			}
		}
	}
	return
}
