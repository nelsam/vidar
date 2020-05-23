// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/editor"
)

type BindManager interface {
	Bindable(string) bind.Bindable
	Execute(bind.Bindable)
}

type EditorChooser interface {
	NextEditor(editor.Direction) text.Editor
}

type Locationer interface {
	bind.Bindable
	For(...focus.Opt) bind.Bindable
}

type ChangeFocus struct {
	direction editor.Direction
	name      string

	binder  BindManager
	chooser EditorChooser
}

func NewFocusRight() *ChangeFocus {
	return &ChangeFocus{direction: editor.Right, name: "focus-right"}
}

func NewFocusLeft() *ChangeFocus {
	return &ChangeFocus{direction: editor.Left, name: "focus-left"}
}

func NewFocusUp() *ChangeFocus {
	return &ChangeFocus{direction: editor.Up, name: "focus-up"}
}

func NewFocusDown() *ChangeFocus {
	return &ChangeFocus{direction: editor.Down, name: "focus-down"}
}

func (p *ChangeFocus) Name() string {
	return p.name
}

func (p *ChangeFocus) Menu() string {
	return "View"
}

func (p *ChangeFocus) Defaults() []fmt.Stringer {
	e := gxui.KeyboardEvent{
		Modifier: gxui.ModAlt,
	}
	switch p.direction {
	case editor.Right:
		e.Key = gxui.KeyRight
	case editor.Left:
		e.Key = gxui.KeyLeft
	case editor.Up:
		e.Key = gxui.KeyUp
	case editor.Down:
		e.Key = gxui.KeyDown
	default:
		panic(fmt.Errorf("Direction %d is invalid", p.direction))
	}
	return []fmt.Stringer{e}
}

func (p *ChangeFocus) Reset() {
	p.binder = nil
	p.chooser = nil
}

func (p *ChangeFocus) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case BindManager:
		p.binder = src
	case EditorChooser:
		p.chooser = src
	}
	if p.chooser != nil && p.binder != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (p *ChangeFocus) Exec() error {
	opener, ok := p.binder.Bindable("focus-location").(Locationer)
	if !ok {
		return errors.New("no open-file command found of type Opener")
	}
	editor := p.chooser.NextEditor(p.direction)
	if editor == nil {
		return errors.New("no editor to switch to")
	}
	p.binder.Execute(opener.For(focus.Path(editor.Filepath())))
	return nil
}
