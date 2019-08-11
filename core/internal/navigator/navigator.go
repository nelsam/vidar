// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"github.com/nelsam/gxui"

	// TODO: Remove once we're fully decoupled.
	vgxui "github.com/nelsam/vidar/core/gxui"
	"github.com/nelsam/vidar/ui"
)

// Pane is a type that has a button and a window frame.
type Pane interface {
	// Button returns the button that is shown for displaying the
	// Pane's Frame.
	Button() gxui.Button

	// Frame returns the frame that is displayed when the Pane's
	// Button is clicked.
	Frame() gxui.Control
}

// HeightSetter is any type which needs its height set explicitly
type HeightSetter interface {
	SetHeight(int)
}

type Creator interface {
	Runner() ui.Runner
	LinearLayout(ui.Direction) (ui.Layout, error)
}

// Navigator is a type implementing the navigation pane of vidar.
type Navigator struct {
	runner ui.Runner
	layout ui.Layout

	buttons gxui.LinearLayout
	frame   gxui.Control

	panes []Pane
}

// New creates and returns a new *Navigator and its UI layout.
func New(creator Creator) (*Navigator, ui.Layout, error) {
	layout, err := creator.LinearLayout(ui.Left)
	if err != nil {
		return nil, nil, err
	}
	nav := &Navigator{
		runner: creator.Runner(),
		layout: layout,
	}

	driver, theme := creator.(*vgxui.Creator).Raw()
	nav.buttons = theme.CreateLinearLayout()
	nav.buttons.SetDirection(gxui.TopToBottom)
	nav.layout.Add(vgxui.Control{nav.buttons})

	return nav, layout, nil
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
		if event.Button != gxui.MouseButtonLeft {
			return
		}
		n.ToggleNavPane(pane.Frame())
	})
	n.runner.Enqueue(func() {
		n.buttons.AddChild(button)
	})
}

func (n *Navigator) ToggleNavPane(frame gxui.Control) {
	if n.frame == frame {
		n.HideNavPane()
		return
	}
	n.ShowNavPane(frame)
}

func (n *Navigator) Resize(height int) {
	for _, pane := range n.panes {
		if heightSetter, ok := pane.(HeightSetter); ok {
			heightSetter.SetHeight(height)
		}
	}
}

func (n *Navigator) HideNavPane() {
	if n.frame == nil {
		return
	}
	n.layout.RemoveChild(n.frame)
	n.frame = nil
}

func (n *Navigator) ShowNavPane(frame gxui.Control) {
	n.HideNavPane()
	if frame == nil {
		return
	}
	n.frame = frame
	n.layout.AddChild(n.frame)
	if focusable, ok := n.frame.(gxui.Focusable); ok {
		gxui.SetFocus(focusable)
	}
}
