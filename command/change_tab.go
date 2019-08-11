// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/ui"
)

type TabChooser interface {
	EditorAt(ui.Direction) input.Editor
}

type ChangeTab struct {
	shift ui.Direction
	name  string

	binder  BindManager
	chooser TabChooser
}

func NewNextTab() *ChangeTab {
	return &ChangeTab{shift: ui.Right, name: "next-tab"}
}

func NewPrevTab() *ChangeTab {
	return &ChangeTab{shift: ui.Left, name: "prev-tab"}
}

func (t *ChangeTab) Name() string {
	return t.name
}

func (t *ChangeTab) Menu() string {
	return "View"
}

func (t *ChangeTab) Defaults() []fmt.Stringer {
	e := gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyTab,
	}
	if t.shift == ui.Left {
		e.Modifier |= gxui.ModShift
	}
	return []fmt.Stringer{e}
}

func (t *ChangeTab) Reset() {
	t.binder = nil
	t.chooser = nil
}

func (t *ChangeTab) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case BindManager:
		t.binder = src
	case TabChooser:
		t.chooser = src
	}
	if t.chooser != nil && t.binder != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (t *ChangeTab) Exec() error {
	opener, ok := t.binder.Bindable("focus-location").(Locationer)
	if !ok {
		return errors.New("no open-file command found of type Opener")
	}
	editor := t.chooser.EditorAt(t.shift)
	if editor == nil {
		return errors.New("no editor to switch to")
	}
	t.binder.Execute(opener.For(focus.Path(editor.Filepath())))
	return nil
}
