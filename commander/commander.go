package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
)

type Commander struct {
	mixins.LinearLayout

	driver gxui.Driver
	theme  *basic.Theme
	file   string
}

func New(driver gxui.Driver, theme *basic.Theme) *Commander {
	commander := &Commander{
		driver: driver,
		theme:  theme,
	}
	commander.LinearLayout.Init(commander, theme)
	commander.SetDirection(gxui.LeftToRight)
	size := theme.DefaultMonospaceFont().GlyphMaxSize()
	size.W = math.MaxSize.W
	commander.SetSize(size)
	return commander
}

func (c *Commander) Clear() {
	c.RemoveAll()
}

func (c *Commander) CurrentFile(file string) {
	c.Clear()

	c.file = file
	box := c.theme.CreateLabel()
	box.SetText(file)
	c.AddChild(box)
}

func (c *Commander) PromptOpenFile(callback func(file string)) {
	c.Clear()

	label := c.theme.CreateLabel()
	label.SetText("Open File:")
	c.AddChild(label)

	commanderCallback := func(file string) {
		c.CurrentFile(file)
		callback(file)
	}
	file := NewFSLocator(c.driver, c.theme, c.file, commanderCallback)
	c.AddChild(file)
	gxui.SetFocus(file)
}
