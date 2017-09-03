// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"log"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins/base"
	"github.com/nelsam/gxui/mixins/parts"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/controller"
	"github.com/nelsam/vidar/settings"
)

// Controller is a type which is used by the Commander to control the
// main UI.
type Controller interface {
	gxui.Control
	Editor() controller.Editor
}

// Commander is a gxui.LinearLayout that takes care of displaying the
// command utilities around a controller.
type Commander struct {
	base.Container
	parts.BackgroundBorderPainter

	driver gxui.Driver
	theme  *basic.Theme

	controller Controller
	box        *commandBox

	inputHandler input.Handler

	lock sync.RWMutex

	stack    []map[string]bind.Bindable
	commands map[gxui.KeyboardEvent]bind.Command
	menuBar  *menuBar
}

// New creates and initializes a *Commander, then returns it.
func New(driver gxui.Driver, theme *basic.Theme, controller Controller) *Commander {
	commander := &Commander{
		driver: driver,
		theme:  theme,
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
	commander.box = newCommandBox(driver, theme, commander.controller)

	mainLayout.AddChild(commander.menuBar)

	subLayout := theme.CreateLinearLayout()
	subLayout.SetDirection(gxui.BottomToTop)
	subLayout.AddChild(commander.box)
	subLayout.AddChild(commander.controller)
	mainLayout.AddChild(subLayout)
	commander.AddChild(mainLayout)
	return commander
}

func (c *Commander) InputHandler() input.Handler {
	return c.inputHandler
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

func (c *Commander) cloneTop() map[string]bind.Bindable {
	m := make(map[string]bind.Bindable)
	if len(c.stack) == 0 {
		return m
	}
	for n, b := range c.stack[len(c.stack)-1] {
		if h, ok := b.(input.Handler); ok {
			b = h.New()
		}
		m[n] = b
	}
	return m
}

// Push binds all bindables to c, pushing the previous
// binding down in the stack.
func (c *Commander) Push(bindables ...bind.Bindable) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.menuBar.Clear()
	defer c.mapMenu()

	next := c.cloneTop()
	c.stack = append(c.stack, next)
	c.commands = make(map[gxui.KeyboardEvent]bind.Command)
	defer c.mapBindings()

	var hooks []bind.OpHook
	for _, b := range bindables {
		next[b.Name()] = b
		if h, ok := b.(bind.OpHook); ok {
			hooks = append(hooks, h)
		}
	}

	multiEditor := c.controller.Editor()
	if multiEditor != nil {
		e := multiEditor.CurrentEditor()
		if e != nil {
			defer c.driver.Call(func() {
				c.inputHandler.Init(e, e.Controller().TextRunes())
			})
		}
	}

	for _, h := range hooks {
		b := next[h.OpName()]
		if b == nil {
			log.Printf("Warning: binding %s (requested by hook %s) is not found", h.OpName(), h.Name())
			continue
		}
		var (
			newOp bind.Bindable
			err   error
		)
		switch op := b.(type) {
		case bind.HookedOp:
			newOp, err = op.Bind(h)
		case bind.HookedMultiOp:
			newOp, err = op.Bind(h)
		case input.Handler:
			newOp, err = op.Bind(h)
		default:
			log.Printf("Warning: binding %s (requested by hook %s) is not a HookedOp and cannot bind hooks to itself. Skipping.", b.Name(), h.Name())
			continue
		}
		if err != nil {
			log.Printf("Warning: failed to bind hook %s to op %s: %s", h.Name(), b.Name(), err)
			continue
		}
		next[newOp.Name()] = newOp
	}
}

func (c *Commander) mapBindings() {
	var handler input.Handler
	for _, b := range c.stack[len(c.stack)-1] {
		switch src := b.(type) {
		case bind.Command:
			c.bind(src, settings.Bindings(src.Name())...)
		case input.Handler:
			handler = src
		}
	}
	if handler == nil {
		log.Fatal("There is no input handler available!  This should never happen.  Please create an issue in github stating that you saw this message.")
	}
	c.inputHandler = handler
}

func (c *Commander) mapMenu() {
	keys := make(map[bind.Command][]gxui.KeyboardEvent)
	var cmds []bind.Command
	for key, bound := range c.commands {
		if _, ok := keys[bound]; !ok {
			cmds = append(cmds, bound)
		}
		keys[bound] = append(keys[bound], key)
	}
	for _, cmd := range cmds {
		c.menuBar.Add(cmd, keys[cmd]...)
	}
}

// Pop pops the most recent call to Bind, restoring the
// previous bind.
func (c *Commander) Pop() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.menuBar.Clear()
	defer c.mapMenu()

	c.commands = make(map[gxui.KeyboardEvent]bind.Command)
	defer c.mapBindings()

	c.stack = c.stack[:len(c.stack)-1]
}

func (c *Commander) bind(command bind.Command, bindings ...gxui.KeyboardEvent) {
	for _, binding := range bindings {
		if old, ok := c.commands[binding]; ok {
			log.Printf("Warning: command %s is overriding command %s at binding %v", command.Name(), old.Name(), binding)
		}
		c.commands[binding] = command
	}
}

// Binding finds and returns the Command associated with bind.
func (c *Commander) Binding(binding gxui.KeyboardEvent) bind.Command {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.commands[binding]
}

// Command looks up a bind.Command by name.
func (c *Commander) Command(name string) bind.Command {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.command(name)
}

func (c *Commander) command(name string) bind.Command {
	if cmd, ok := c.stack[len(c.stack)-1][name].(bind.Command); ok {
		return cmd
	}
	return nil
}

// KeyPress handles key bindings for c.
func (c *Commander) KeyPress(event gxui.KeyboardEvent) (consume bool) {
	editor := c.controller.Editor()
	if event.Modifier == 0 && event.Key == gxui.KeyEscape {
		c.box.Clear()
		editor.Focus()
		return true
	}
	codeEditor := editor.CurrentEditor()
	if codeEditor != nil && codeEditor.HasFocus() {
		c.inputHandler.HandleEvent(codeEditor, event)
	}
	if command := c.Binding(event); command != nil {
		c.box.Clear()
		if c.box.Run(command) {
			return true
		}
		c.Execute(c.box.Current())
		c.box.Finish()
		return true
	}
	if !c.box.HasFocus() {
		return false
	}
	if c.box.Finished(event) {
		c.Execute(c.box.Current())
		c.box.Finish()
	}
	return true
}

func (c *Commander) KeyStroke(event gxui.KeyStrokeEvent) (consume bool) {
	if event.Modifier&^gxui.ModShift != 0 {
		return false
	}
	e := c.controller.Editor().CurrentEditor()
	if e == nil || !e.HasFocus() {
		return false
	}
	c.inputHandler.HandleInput(e, event)
	return true
}

func (c *Commander) Execute(e bind.Command) {
	if before, ok := e.(BeforeExecutor); ok {
		before.BeforeExec(c)
	}
	var status bind.Status
	switch src := e.(type) {
	case bind.Op:
		status = execute(src.Exec, c)
	case bind.MultiOp:
		src.Reset()
		status = execute(src.Store, c)
		if status&bind.Executed == 0 {
			break
		}
		if err := src.Exec(); err != nil {
			log.Printf("Error executing multi-executor: %s", err)
		}
	default:
		return
	}
	if status&bind.Executed == 0 {
		log.Printf("Warning: Executor of type %T ran without executing", e)
	}
}

func (c *Commander) Elements() []interface{} {
	all := make([]interface{}, 0, len(c.commands))
	for _, binding := range c.commands {
		all = append(all, binding)
	}
	all = append(all, c.controller, c.inputHandler)
	return all
}

func execute(executor func(interface{}) bind.Status, elem interface{}) bind.Status {
	status := executor(elem)
	if status&bind.Stop != 0 {
		return status
	}

	switch src := elem.(type) {
	case Elementer:
		for _, element := range src.Elements() {
			status |= execute(executor, element)
			if status&bind.Stop != 0 {
				break
			}
		}
	case gxui.Parent:
		for _, child := range src.Children() {
			status |= execute(executor, child.Control)
			if status&bind.Stop != 0 {
				break
			}
		}
	}
	return status
}
