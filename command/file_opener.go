// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/status"
)

type Focuser interface {
	For(...focus.Opt) bind.Bindable
}

type FileOpener struct {
	status.General

	driver gxui.Driver
	theme  *basic.Theme

	file  *FSLocator
	input <-chan gxui.Focusable

	focuser Focuser
	execer  Executor
}

func NewFileOpener(driver gxui.Driver, theme *basic.Theme) *FileOpener {
	o := &FileOpener{
		driver: driver,
		theme:  theme,
	}
	o.file = NewFSLocator(driver, theme)
	return o
}

func (f *FileOpener) Name() string {
	return "open-file"
}

func (f *FileOpener) Menu() string {
	return "File"
}

func (f *FileOpener) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyO,
	}}
}

func (f *FileOpener) Start(control gxui.Control) gxui.Control {
	f.file.LoadDir(control)
	input := make(chan gxui.Focusable, 1)
	f.input = input
	input <- f.file
	close(input)
	return nil
}

func (f *FileOpener) Next() gxui.Focusable {
	return <-f.input
}

func (f *FileOpener) Reset() {
	f.focuser = nil
	f.execer = nil
}

func (f *FileOpener) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case Focuser:
		f.focuser = src
	case Executor:
		f.execer = src
	}

	if f.focuser == nil || f.execer == nil {
		return bind.Waiting
	}
	return bind.Executing
}

func (f *FileOpener) Exec() error {
	path := f.file.Path()
	if path == "" {
		return errors.New("command.FileOpener: No file path provided")
	}
	f.execer.Execute(f.focuser.For(focus.Path(path)))
	return nil
}
