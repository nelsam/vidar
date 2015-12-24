// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/controller"
)

// Completer is a type which defines when a gxui.KeyboardEvent
// completes an action.  Types returned from
// "../commands".Command.Next() may implement this if they don't want
// to immediately complete when the "enter" key is pressed.
type Completer interface {
	// Complete returns whether or not the key signals a completion of
	// the input.
	Complete(gxui.KeyboardEvent) bool
}

// Controller is a type which controls the main editor UI.
type Controller interface {
	gxui.Control
	Execute(commands.Command)
	Editor() controller.Editor
	Navigator() controller.Navigator
}

// Commander is a gxui.LinearLayout that includes a Controller (taking
// up the majority of the Commander) and a CommandBox (taking up one
// line at the bottom of the Commander).  Commands can be manually
// entered through the command box or bound to keyboard shortcuts.
type Commander struct {
	mixins.LinearLayout

	driver     gxui.Driver
	theme      *basic.Theme
	font       gxui.Font
	controller Controller
	box        *commandBox

	commands []commands.Command
	bindings map[gxui.KeyboardEvent]commands.Command
}

// New creates and initializes a *Commander, then returns it.
func New(driver gxui.Driver, theme *basic.Theme, font gxui.Font) *Commander {
	commander := new(Commander)
	commander.Init(driver, theme, font)
	return commander
}

// Init resets c's state.
func (c *Commander) Init(driver gxui.Driver, theme *basic.Theme, font gxui.Font) {
	c.LinearLayout.Init(c, theme)
	c.bindings = make(map[gxui.KeyboardEvent]commands.Command)
	c.driver = driver
	c.theme = theme
	c.font = font
	c.SetDirection(gxui.BottomToTop)
	c.SetSize(math.MaxSize)

	c.controller = controller.New(c.driver, c.theme, c.font)
	c.box = newCommandBox(c.theme, c.controller)

	c.AddChild(c.box)
	c.AddChild(c.controller)

	// TODO: Store these in a config file or something
	openFile := commands.NewFileOpener(c.driver, c.theme)
	c.commands = append(c.commands, openFile)
	ctrlO := gxui.KeyboardEvent{
		Key:      gxui.KeyO,
		Modifier: gxui.ModControl,
	}
	supO := gxui.KeyboardEvent{
		Key:      gxui.KeyO,
		Modifier: gxui.ModSuper,
	}
	c.bindings[ctrlO] = openFile
	c.bindings[supO] = openFile

	addProject := commands.NewProjectAdder(c.driver, c.theme)
	c.commands = append(c.commands, addProject)
	ctrlShiftN := gxui.KeyboardEvent{
		Key:      gxui.KeyN,
		Modifier: gxui.ModControl | gxui.ModShift,
	}
	supShiftN := gxui.KeyboardEvent{
		Key:      gxui.KeyN,
		Modifier: gxui.ModSuper | gxui.ModShift,
	}
	c.bindings[ctrlShiftN] = addProject
	c.bindings[supShiftN] = addProject

	find := commands.NewFinder(c.driver, c.theme)
	ctrlF := gxui.KeyboardEvent{
		Key:      gxui.KeyF,
		Modifier: gxui.ModControl,
	}
	supF := gxui.KeyboardEvent{
		Key:      gxui.KeyF,
		Modifier: gxui.ModSuper,
	}
	c.bindings[ctrlF] = find
	c.bindings[supF] = find
}

// Controller returns c's controller.
func (c *Commander) Controller() Controller {
	return c.controller
}

// KeyPress handles key bindings for c.
func (c *Commander) KeyPress(event gxui.KeyboardEvent) (consume bool) {
	if event.Modifier == 0 && event.Key == gxui.KeyEscape {
		c.box.Clear()
		c.controller.Editor().Focus()
		return true
	}
	cmdDone := c.box.HasFocus() && c.box.Done()
	if command, ok := c.bindings[event]; ok {
		c.box.Clear()
		if c.box.Run(command) {
			return true
		}
		cmdDone = true
	}
	if cmdDone {
		c.controller.Execute(c.box.Current())
		c.box.Clear()
		return true
	}
	return false
}
