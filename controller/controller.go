// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package controller

import (
	"go/token"
	"log"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
)

type Navigator interface {
	gxui.Control
}

type Editor interface {
	gxui.Control
	Focus()
	CurrentFile() string
	Open(path string, cursor token.Position)
}

type Elementer interface {
	Elements() []interface{}
}

// An Executor is a type which can execute operations on the
// controller and/or its children.
type Executor interface {
	// Exec executes the operation(s).  Exec will be called once for
	// the controller and each of its children, until consume is
	// true.  If Exec has been run against the controller and all its
	// children without ever returning true for executed, it is
	// assumed to be unexpected behavior, and an error will be shown.
	Exec(interface{}) (executed, consume bool)
}

// A BeforeExecutor is an Executor which has tasks to run on the
// Controller prior to running Exec.
type BeforeExecutor interface {
	BeforeExec(interface{})
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

// Execute implements "../commander".Controller.
func (c *Controller) Execute(executor Executor) {
	if before, ok := executor.(BeforeExecutor); ok {
		before.BeforeExec(c)
	}
	executed, _ := execRecursively(executor, c)
	if !executed {
		log.Printf("Warning: Executor of type %T ran without executing", executor)
	}
	c.Editor().Focus()
}

func execRecursively(executor Executor, element interface{}) (executed, consume bool) {
	executed, consume = executor.Exec(element)
	if consume {
		return
	}
	var childExecuted bool
	if parent, ok := element.(gxui.Parent); ok {
		for _, child := range parent.Children() {
			childExecuted, consume = execRecursively(executor, child.Control)
			executed = executed || childExecuted
			if consume {
				return
			}
		}
	}
	if elementer, ok := element.(Elementer); ok {
		for _, element := range elementer.Elements() {
			childExecuted, consume = execRecursively(executor, element)
			executed = executed || childExecuted
			if consume {
				return
			}
		}
	}
	return
}
