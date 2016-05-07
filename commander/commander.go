// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins/base"
	"github.com/nelsam/gxui/mixins/parts"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/controller"
)

// Controller is a type which is used by the Commander to control the
// main UI.
type Controller interface {
	gxui.Control
	Execute(controller.Executor)
	Editor() controller.Editor
	Navigator() controller.Navigator
}

// A commandMapping is a mapping between keyboard shortcuts (if any),
// a menu name, and a command.  The menu name is required.
type commandMapping struct {
	binding gxui.KeyboardEvent
	command Command
}

// Commander is a gxui.LinearLayout that takes care of displaying the
// command utilities around a controller.
type Commander struct {
	base.Container
	parts.BackgroundBorderPainter

	theme *basic.Theme

	controller Controller
	box        *commandBox

	commands []commandMapping
	menuBar  *menuBar
}

// New creates and initializes a *Commander, then returns it.
func New(theme *basic.Theme, controller Controller) *Commander {
	commander := &Commander{
		theme: theme,
	}
	commander.Container.Init(commander, theme)
	commander.BackgroundBorderPainter.Init(commander)
	commander.SetMouseEventTarget(true)
	commander.SetBackgroundBrush(gxui.TransparentBrush)
	commander.SetBorderPen(gxui.TransparentPen)

	mainLayout := theme.CreateLinearLayout()

	mainLayout.SetDirection(gxui.TopToBottom)
	mainLayout.SetSize(math.MaxSize)

	commander.controller = controller
	commander.menuBar = newMenuBar(commander, theme)
	commander.box = newCommandBox(theme, commander.controller)

	mainLayout.AddChild(commander.menuBar)

	subLayout := theme.CreateLinearLayout()
	subLayout.SetDirection(gxui.BottomToTop)
	subLayout.AddChild(commander.box)
	subLayout.AddChild(commander.controller)
	mainLayout.AddChild(subLayout)
	commander.AddChild(mainLayout)
	return commander
}

func (c *Commander) DesiredSize(_, max math.Size) math.Size {
	return max
}

func (c *Commander) LayoutChildren() {
	maxSize := c.Size().Contract(c.Padding())
	for _, child := range c.Children() {
		margin := child.Control.Margin()
		desiredSize := child.Control.DesiredSize(math.ZeroSize, maxSize.Contract(margin))
		child.Control.SetSize(desiredSize)
	}
}

func (c *Commander) Paint(canvas gxui.Canvas) {
	rect := c.Size().Rect()
	c.BackgroundBorderPainter.PaintBackground(canvas, rect)
	c.PaintChildren.Paint(canvas)
	c.BackgroundBorderPainter.PaintBorder(canvas, rect)
}

// Controller returns c's controller.
func (c *Commander) Controller() Controller {
	return c.controller
}

// Map maps a Command to a menu and an optional key binding.
func (c *Commander) Map(command Command, menu string, bindings ...gxui.KeyboardEvent) error {
	if len(menu) == 0 {
		return fmt.Errorf("All commands must have a menu entry")
	}
	c.menuBar.Add(menu, command, bindings...)
	for _, binding := range bindings {
		if err := c.mapBinding(command, binding); err != nil {
			return err
		}
	}
	return nil
}

func (c *Commander) mapBinding(command Command, binding gxui.KeyboardEvent) error {
	if command := c.Binding(binding); command != nil {
		return fmt.Errorf("Binding %s is already mapped", binding)
	}
	c.commands = append(c.commands, commandMapping{
		binding: binding,
		command: command,
	})
	return nil
}

// Binding finds the Command associated with binding.
func (c *Commander) Binding(binding gxui.KeyboardEvent) Command {
	if binding.Key == gxui.KeyUnknown {
		return nil
	}
	for _, mapping := range c.commands {
		if mapping.binding == binding {
			return mapping.command
		}
	}
	return nil
}

// KeyPress handles key bindings for c.
func (c *Commander) KeyPress(event gxui.KeyboardEvent) (consume bool) {
	if event.Modifier == 0 && event.Key == gxui.KeyEscape {
		c.box.Clear()
		c.controller.Editor().Focus()
		return true
	}
	cmdDone := c.box.HasFocus() && event.Modifier == 0 && event.Key == gxui.KeyEnter
	if command := c.Binding(event); command != nil {
		c.box.Clear()
		if c.box.Run(command) {
			return true
		}
		cmdDone = true
	}
	if !cmdDone {
		return false
	}
	if executor, ok := c.box.Current().(Executor); ok {
		c.controller.Execute(executor)
	}
	c.box.Clear()
	return true
}
