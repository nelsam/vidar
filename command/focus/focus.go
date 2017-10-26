// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package focus

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
)

type EditorOpener interface {
	Open(string) (editor input.Editor, existed bool)
	CurrentEditor() input.Editor
}

type Opener interface {
	Open(string)
}

type LineStarter interface {
	LineStart(int) int
}

type Mover interface {
	To(...int) bind.Bindable
}

// A FileBinder is a type that registers a list of commands
// on file open.
type FileBinder interface {
	// FileBindables returns the bindables that should be registered
	// for the given file.
	FileBindables(path string) []bind.Bindable
}

// A Binder is a type which can bind bindables
type Binder interface {
	Push(...bind.Bindable)
	Pop() []bind.Bindable
	Execute(bind.Bindable)
}

type Opt func(*Location) error

func Path(path string) Opt {
	return func(l *Location) error {
		l.file.SetPath(path)
		return nil
	}
}

func Line(line int) Opt {
	return func(l *Location) error {
		if l.offset >= 0 {
			return errors.New("cannot open a line number and an offset")
		}
		l.line = line
		return nil
	}
}

func Column(col int) Opt {
	return func(l *Location) error {
		if l.offset >= 0 {
			return errors.New("cannot open a column number and an offset")
		}
		l.col = col
		return nil
	}
}

func Offset(offset int) Opt {
	return func(l *Location) error {
		if l.col >= 0 || l.line >= 0 {
			return errors.New("cannot open an offset and a line/column")
		}
		l.offset = offset
		return nil
	}
}

type Location struct {
	status.General

	driver gxui.Driver
	theme  *basic.Theme

	file  *FSLocator
	input <-chan gxui.Focusable

	// These can only be set by FocusOpts in For().
	offset, line, col int

	mover   Mover
	binder  Binder
	opener  EditorOpener
	openers []Opener

	hooks []FileBinder
}

func NewLocation(driver gxui.Driver, theme *basic.Theme) *Location {
	o := &Location{
		driver: driver,
		theme:  theme,
		offset: -1,
		line:   -1,
		col:    -1,
	}
	o.file = NewFSLocator(driver, theme)
	return o
}

func (f *Location) For(opts ...Opt) bind.Bindable {
	newF := NewLocation(f.driver, f.theme)
	newF.file.SetPath(f.file.Path())
	newF.offset = f.offset
	newF.line = f.line
	newF.col = f.col
	newF.hooks = append(newF.hooks, f.hooks...)
	for _, o := range opts {
		if err := o(newF); err != nil {
			if len(newF.Warn) != 0 {
				newF.Warn += "; "
			}
			newF.Warn += fmt.Sprintf("failed to apply option: %s", err)
		}
	}
	return newF
}

func (f *Location) Name() string {
	return "open-file"
}

func (f *Location) Menu() string {
	return "File"
}

func (f *Location) Start(control gxui.Control) gxui.Control {
	f.file.LoadDir(control)
	input := make(chan gxui.Focusable, 1)
	f.input = input
	input <- f.file
	close(input)
	return nil
}

func (f *Location) Next() gxui.Focusable {
	return <-f.input
}

func (f *Location) Reset() {
	f.mover = nil
	f.binder = nil
	f.opener = nil
	f.openers = nil
}

func (f *Location) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case Mover:
		f.mover = src
	case EditorOpener:
		f.opener = src
	case Opener:
		f.openers = append(f.openers, src)
	case Binder:
		f.binder = src
	}

	if !f.moverReady() || f.opener == nil || f.binder == nil {
		return bind.Waiting
	}
	return bind.Executing
}

func (f *Location) moverReady() bool {
	if f.line == -1 && f.col == -1 && f.offset == -1 {
		// The mover is not needed.
		return true
	}
	return f.mover != nil
}

func (f *Location) Exec() error {
	if f.opener.CurrentEditor() != nil {
		f.binder.Pop()
	}
	path := f.file.Path()
	e, _ := f.opener.Open(path)
	for _, o := range f.openers {
		o.Open(path)
	}
	var b []bind.Bindable
	for _, h := range f.hooks {
		b = append(b, h.FileBindables(path)...)
	}
	f.binder.Push(b...)
	f.moveCarets(e.(LineStarter))
	return nil
}

func (f *Location) moveCarets(l LineStarter) {
	if f.offset == -1 && f.line == -1 && f.col == -1 {
		return
	}
	if f.offset != -1 {
		f.binder.Execute(f.mover.To(f.offset))
		return
	}
	offset := 0
	if f.line != -1 {
		offset = l.LineStart(f.line)
	}
	if f.col != -1 {
		offset += f.col
	}
	f.binder.Execute(f.mover.To(offset))
}

func (f *Location) Bind(h bind.Bindable) (bind.HookedMultiOp, error) {
	bndr, ok := h.(FileBinder)
	if !ok {
		return nil, fmt.Errorf("expected hook to be FileBinder, was %T", h)
	}
	newF := NewLocation(f.driver, f.theme)
	newF.hooks = append(newF.hooks, f.hooks...)
	newF.hooks = append(newF.hooks, bndr)
	return newF, nil
}
