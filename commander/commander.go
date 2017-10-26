// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"fmt"
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
	"github.com/nelsam/vidar/setting"
)

// Controller is a type which is used by the Commander to control the
// main UI.
type Controller interface {
	gxui.Control
	Editor() controller.MultiEditor
}

type Controllable interface {
	Controller() *gxui.TextBoxController
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

	stack    [][]bind.Bindable
	bound    map[string]bind.Bindable
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

func (c *Commander) cloneTop() []bind.Bindable {
	if len(c.stack) == 0 {
		return nil
	}
	next := make([]bind.Bindable, 0, len(c.stack[len(c.stack)-1]))
	for _, b := range c.stack[len(c.stack)-1] {
		if h, ok := b.(input.Handler); ok {
			// TODO: decide if this is helpful.  Bind now returns a new
			// input.Handler each time, so this is probably useless.
			b = h.New()
		}
		next = append(next, b)
	}
	return next
}

// Push binds all bindables to c, pushing the previous
// binding down in the stack.
func (c *Commander) Push(bindables ...bind.Bindable) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.menuBar.Clear()
	defer c.mapMenu()

	c.stack = append(c.stack, append(c.cloneTop(), bindables...))
	c.commands = make(map[gxui.KeyboardEvent]bind.Command)
	defer c.mapBindings()

	c.bindStack()
}

func (c *Commander) bindStack() {
	c.bound = make(map[string]bind.Bindable, len(c.stack[len(c.stack)-1]))
	var (
		hooks      []bind.OpHook
		multiHooks []bind.MultiOpHook
	)
	for _, b := range c.stack[len(c.stack)-1] {
		c.bound[b.Name()] = b
		switch h := b.(type) {
		case bind.OpHook:
			hooks = append(hooks, h)
		case bind.MultiOpHook:
			multiHooks = append(multiHooks, h)
		}
	}

	if e := c.editor(c.controller.Editor()); e != nil {
		defer c.driver.Call(func() {
			c.inputHandler.Init(e, e.(Controllable).Controller().TextRunes())
		})
	}

	for _, h := range hooks {
		bindNames(c.bound, h, h.OpName())
	}
	for _, h := range multiHooks {
		bindNames(c.bound, h, h.OpNames()...)
	}
}

func (c *Commander) editor(e controller.MultiEditor) input.Editor {
	if e == nil {
		return nil
	}
	return e.CurrentEditor()
}

func (c *Commander) mapBindings() {
	var handler input.Handler
	// Loop through the slice to preserve order.
	for _, b := range c.stack[len(c.stack)-1] {
		// Load from the map to ensure hooks are bound.
		b = c.bound[b.Name()]
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
	for key, bound := range c.commands {
		keys[bound] = append(keys[bound], key)
	}
	// As usual, use the stack slice to preserve order
	for _, b := range c.stack[len(c.stack)-1] {
		cmd, ok := b.(bind.Command)
		if !ok {
			continue
		}
		c.menuBar.Add(cmd, keys[cmd]...)
	}
}

// Pop pops the most recent call to Bind, restoring the
// previous bind.
func (c *Commander) Pop() []bind.Bindable {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.menuBar.Clear()
	defer c.mapMenu()

	c.commands = make(map[gxui.KeyboardEvent]bind.Command)
	defer c.mapBindings()

	end := len(c.stack) - 1
	top := c.stack[end]
	c.stack = c.stack[:end]

	c.bindStack()
	return top
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

// Bindable looks up a bind.Bindable by name
func (c *Commander) Bindable(name string) bind.Bindable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if b, ok := c.bound[name]; ok {
		return b
	}
	return nil
}

// KeyPress handles key bindings for c.
func (c *Commander) KeyPress(event gxui.KeyboardEvent) (consume bool) {
	editor := c.controller.Editor()
	if event.Modifier == 0 && event.Key == gxui.KeyEscape {
		c.box.Clear()
		if e := editor.CurrentEditor(); e != nil {
			gxui.SetFocus(e.(gxui.Focusable))
		}
	}
	codeEditor := editor.CurrentEditor()
	if codeEditor != nil && codeEditor.(gxui.Focusable).HasFocus() {
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
	if e == nil || !e.(gxui.Focusable).HasFocus() {
		return false
	}
	c.inputHandler.HandleInput(e, event)
	return true
}

func (c *Commander) Execute(e bind.Bindable) {
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
	all := make([]interface{}, 0, len(c.bound))
	for _, binding := range c.bound {
		all = append(all, binding)
	}
	all = append(all, c.controller, c.inputHandler)
	return all
}

func bindNames(m map[string]bind.Bindable, h bind.Bindable, opNames ...string) {
	for _, name := range opNames {
		b := m[name]
		if b == nil {
			log.Printf("Warning: binding %s (requested by hook %s) is not found", name, h.Name())
			continue
		}
		newOp, err := bindName(b, h)
		if err != nil {
			log.Printf("Warning: failed to bind hook %s to op %s: %s", h.Name(), b.Name(), err)
			continue
		}
		m[newOp.Name()] = newOp
	}
}

func bindName(b, h bind.Bindable) (bind.Bindable, error) {
	switch op := b.(type) {
	case bind.HookedOp:
		return op.Bind(h)
	case bind.HookedMultiOp:
		return op.Bind(h)
	case input.Handler:
		return op.Bind(h)
	default:
		return nil, fmt.Errorf("binding %s (requested by hook %s) is not a HookedOp and cannot bind hooks to itself. Skipping.", b.Name(), h.Name())
	}
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
