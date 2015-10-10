package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
)

type Commander struct {
	gxui.LinearLayout

	theme gxui.Theme
	file  string
}

func New(theme gxui.Theme) *Commander {
	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	size := theme.DefaultMonospaceFont().GlyphMaxSize()
	size.W = math.MaxSize.W
	layout.SetSize(size)
	return &Commander{
		LinearLayout: layout,
		theme:        theme,
	}
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

	file := c.theme.CreateTextBox()
	file.SetDesiredWidth(math.MaxSize.W)
	file.SetMultiline(false)
	file.SetText(c.file)
	file.OnKeyPress(func(event gxui.KeyboardEvent) {
		if event.Key == gxui.KeyEnter {
			callback(file.Text())
			c.CurrentFile(file.Text())
		}
	})
	c.AddChild(file)

	gxui.SetFocus(file)
}
