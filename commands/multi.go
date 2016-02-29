package commands

import "github.com/nelsam/gxui"

type Command interface {
	Start(gxui.Control) gxui.Control
	Name() string
	Next() gxui.Focusable
	Exec(interface{}) (executed, consume bool)
}

type MultiCommand struct {
	theme          gxui.Theme
	display        gxui.LinearLayout
	currentDisplay gxui.Control
	control        gxui.Control

	all      []Command
	needExec []Command
	executed []bool
	current  Command
	incoming chan Command

	prev []interface{}
}

func NewMulti(theme gxui.Theme, commands ...Command) *MultiCommand {
	if len(commands) == 0 {
		panic("commands.NewMulti called without any commands")
	}
	return &MultiCommand{
		theme: theme,
		all:   commands,
	}
}

func (c *MultiCommand) Start(control gxui.Control) gxui.Control {
	c.display = c.theme.CreateLinearLayout()
	c.needExec = append([]Command{}, c.all...)
	c.executed = make([]bool, len(c.all))
	c.incoming = make(chan Command, len(c.all))
	for _, cmd := range c.all {
		c.incoming <- cmd
	}
	close(c.incoming)
	c.nextCommand()
	return c.display
}

func (c *MultiCommand) nextCommand() {
	if c.currentDisplay != nil {
		c.display.RemoveChild(c.currentDisplay)
	}
	c.current = <-c.incoming
	if c.current == nil {
		return
	}
	c.currentDisplay = c.current.Start(c.control)
	if c.currentDisplay != nil {
		c.display.AddChild(c.currentDisplay)
	}
}

func (c *MultiCommand) Name() string {
	name := ""
	for _, cmd := range c.all {
		if name != "" {
			name += ", "
		}
		name += cmd.Name()
	}
	return name
}

func (c *MultiCommand) Next() gxui.Focusable {
	if c.current == nil {
		return nil
	}
	currentNext := c.current.Next()
	if currentNext == nil {
		c.nextCommand()
		return c.Next()
	}
	return currentNext
}

func (c *MultiCommand) Exec(target interface{}) (executed, consume bool) {
	allExecuted := true
	for i := 0; i < len(c.needExec); i++ {
		executed, consume := c.needExec[i].Exec(target)
		if consume {
			c.needExec = append(c.needExec[:i], c.needExec[i+1:]...)
			c.executed = append(c.executed[:i], c.executed[i+1:]...)
			i--
			continue
		}
		if executed {
			c.executed[i] = true
		}
		if !c.executed[i] {
			allExecuted = false
		}
	}
	return allExecuted, len(c.needExec) == 0
}
