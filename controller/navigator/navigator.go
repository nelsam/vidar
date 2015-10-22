// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package navigator

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/settings"
)

// Pane is a type that has a button and a window frame.
type Pane interface {
	// Button returns the button that is shown for displaying the
	// Pane's Frame.
	Button() gxui.Button

	// Frame returns the frame that is displayed when the Pane's
	// Button is clicked.
	Frame() gxui.Focusable
}

type Navigator struct {
	mixins.LinearLayout

	theme   *basic.Theme
	buttons gxui.LinearLayout
	frame   gxui.Focusable

	panes []Pane
}

func New(driver gxui.Driver, theme *basic.Theme, font gxui.Font) *Navigator {
	nav := new(Navigator)
	nav.Init(driver, theme)
	return nav
}

func (n *Navigator) Init(driver gxui.Driver, theme *basic.Theme) {
	n.LinearLayout.Init(n, theme)
	n.theme = theme
	n.SetDirection(gxui.LeftToRight)

	n.buttons = theme.CreateLinearLayout()
	n.buttons.SetDirection(gxui.TopToBottom)
	n.AddChild(n.buttons)

	projects := new(Projects)
	projects.Init(driver, theme, settings.AssetsDir)
	n.Add(projects)

	dirs := new(Directories)
	dirs.Init(driver, theme, settings.AssetsDir)
	n.Add(dirs)
}

func (n *Navigator) Elements() []interface{} {
	elements := make([]interface{}, 0, len(n.panes))
	for _, pane := range n.panes {
		elements = append(elements, pane)
	}
	return elements
}

func (n *Navigator) Buttons() gxui.LinearLayout {
	return n.buttons
}

func (n *Navigator) Add(pane Pane) {
	n.panes = append(n.panes, pane)
	button := pane.Button()
	button.OnClick(func(event gxui.MouseEvent) {
		if n.frame != nil {
			disable := n.frame == pane.Frame()
			n.RemoveChild(n.frame)
			n.frame = nil
			if disable {
				return
			}
		}
		n.frame = pane.Frame()
		if n.frame != nil {
			n.AddChild(n.frame)
			gxui.SetFocus(n.frame)
		}
	})
	n.buttons.AddChild(button)
}
