// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/vidar/controller"
)

// Pane is a type that has a button and a window frame.
type Pane interface {
	// Button returns the button that is shown for displaying the
	// Pane's Frame.
	Button() gxui.Button

	// Frame returns the frame that is displayed when the Pane's
	// Button is clicked.
	Frame() gxui.Control

	// OnComplete takes a function which takes a command, and runs
	// that function when the Pane's action is complete.
	OnComplete(func(controller.Executor))
}

// CommandExecutor is a type that can execute a commands.Command
type CommandExecutor interface {
	Execute(controller.Executor)
}

// Caller is any type that can call a function on the UI goroutine
type Caller interface {
	Call(func()) bool
}

// HeightSetter is any type which needs its height set explicitly
type HeightSetter interface {
	SetHeight(int)
}

// Navigator is a type implementing the navigation pane of vidar.
type Navigator struct {
	mixins.LinearLayout

	caller Caller

	buttons gxui.LinearLayout
	frame   gxui.Control

	panes []Pane

	executor CommandExecutor
}

// New creates and returns a new *Navigator.
func New(driver gxui.Driver, theme gxui.Theme, executor CommandExecutor) *Navigator {
	nav := &Navigator{}
	nav.Init(nav, theme)

	nav.SetDirection(gxui.LeftToRight)
	nav.caller = driver
	nav.executor = executor

	nav.buttons = theme.CreateLinearLayout()
	// TODO: update buttons to use more restrictive type
	nav.buttons.SetDirection(gxui.TopToBottom)
	nav.AddChild(nav.buttons)

	return nav
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
	pane.OnComplete(n.executor.Execute)
	n.panes = append(n.panes, pane)
	button := pane.Button()
	button.OnClick(func(event gxui.MouseEvent) {
		if event.Button != gxui.MouseButtonLeft {
			return
		}
		n.ToggleNavPane(pane.Frame())
	})
	n.caller.Call(func() {
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
	n.RemoveChild(n.frame)
	n.frame = nil
}

func (n *Navigator) ShowNavPane(frame gxui.Control) {
	n.HideNavPane()
	if frame == nil {
		return
	}
	n.frame = frame
	n.AddChild(n.frame)
	if focusable, ok := n.frame.(gxui.Focusable); ok {
		gxui.SetFocus(focusable)
	}
}
