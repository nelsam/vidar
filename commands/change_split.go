// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"errors"

	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/editor"
)

type BindManager interface {
	Bindable(string) bind.Bindable
	Execute(bind.Bindable)
}

type EditorChooser interface {
	NextEditor(editor.Direction) *editor.CodeEditor
}

type Locationer interface {
	bind.Bindable
	For(string, int) bind.Bindable
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
	opener, ok := p.binder.Bindable("open-file").(Locationer)
	if !ok {
		return errors.New("no open-file command found of type Opener")
	}
	editor := p.chooser.NextEditor(p.direction)
	if editor == nil {
		return errors.New("no editor to switch to")
	}
	p.binder.Execute(opener.For(editor.Filepath(), -1))
	return nil
}
