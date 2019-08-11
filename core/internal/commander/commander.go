// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/core/internal/controller"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/setting"
	"github.com/nelsam/vidar/ui"

	// TODO: drop this package from imports when we no longer
	// need access to the core gxui types.
	vgxui "github.com/nelsam/vidar/core/gxui"
)

// Controller is a type which is used by the Commander to control the
// main UI.
type Controller interface {
	gxui.Control
	Editor() controller.MultiEditor
}

type deprecatedControllable interface {
	Controller() *gxui.TextBoxController
}

type Creator interface {
	Runner() ui.Runner
	LinearLayout(ui.Direction) (ui.Layout, error)
	ManualLayout() (ui.Layout, error)
}

// Commander is a gxui.LinearLayout that takes care of displaying the
// command utilities around a controller.
type Commander struct {
	layout ui.Layout

	creator Creator

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
func New(creator Creator, controller Controller) (*Commander, ui.Layout, error) {
	commander := &Commander{
		creator: creator,
	}

	layout, err := creator.ManualLayout()
	if err != nil {
		return nil, nil, err
	}
	commander.layout = layout

	driver, theme := creator.(vgxui.Creator).Raw()

	mainLayout, err := creator.LinearLayout(ui.Up)
	if err != nil {
		return nil, nil, err
	}

	commander.controller = controller
	commander.menuBar = newMenuBar(commander, theme.(*basic.Theme))
	commander.box = newCommandBox(driver, theme, commander.controller)

	mainLayout.Add(vgxui.Control{commander.menuBar})

	subLayout, err := creator.LinearLayout(ui.Down)
	if err != nil {
		return nil, nil, err
	}
	subLayout.Add(vgxui.Control{commander.box})
	subLayout.Add(vgxui.Control{commander.controller})
	mainLayout.Add(subLayout)

	commander.layout.Add(mainLayout)
	return commander, commander.layout, nil
}

func (c *Commander) InputHandler() input.Handler {
	return c.inputHandler
}

func (c *Commander) cloneTop() []bind.Bindable {
	if len(c.stack) == 0 {
		return nil
	}
	current := c.stack[len(c.stack)-1]
	if len(current) == 0 {
		return nil
	}
	next := make([]bind.Bindable, 0, len(current))
	for _, b := range current {
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
		defer c.creator.Runner().Enqueue(func() {
			c.inputHandler.Init(e, e.(deprecatedControllable).Controller().TextRunes())
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
	var (
		handler input.Handler
		cmds    []bind.Command
	)
	// Loop through the slice to preserve order.
	for _, b := range c.stack[len(c.stack)-1] {
		// Load from the map to ensure hooks are bound.
		b = c.bound[b.Name()]
		switch src := b.(type) {
		case bind.Command:
			cmds = append(cmds, src)
		case input.Handler:
			handler = src
		}
	}
	setting.SetDefaultBindings(cmds...)
	for _, cmd := range cmds {
		c.bind(cmd, setting.Bindings(cmd.Name())...)
	}
	if handler == nil {
		log.Fatal("There is no input handler available!  This should never happen.  Please create an issue stating that you saw this message.")
	}
	c.inputHandler = handler
}

func (c *Commander) mapMenu() {
	keys := make(map[string][]gxui.KeyboardEvent)
	for key, bound := range c.commands {
		keys[bound.Name()] = append(keys[bound.Name()], key)
	}
	// As usual, use the stack slice to preserve order
	for _, b := range c.stack[len(c.stack)-1] {
		cmd, ok := b.(bind.Command)
		if !ok {
			continue
		}
		c.menuBar.Add(cmd, keys[cmd.Name()]...)
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

func (c *Commander) Execute(e bind.Bindable) {
	defer func() {
		// Mitigate the potential for plugins to cause the editor to panic
		if r := recover(); r != nil {
			log.Printf("ERR: panic while executing bindable %T: %v", e, r)
			log.Printf("Stack trace:\n%s", debug.Stack())
		}
	}()
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
		log.Printf("Warning: Commander.Execute called against type %T, but it is not a type that can execute.", e)
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
