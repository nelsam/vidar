// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package focus

import (
	"errors"
	"fmt"

	"github.com/nelsam/gxui"
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

// A FileChanger is a type that just needs to be called when
// the open file changes.
type FileChanger interface {
	// FileChanged will be called when the currently focused
	// file changes.
	FileChanged(oldPath, newPath string)
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
		l.path = path
		return nil
	}
}

func Line(line int) Opt {
	return func(l *Location) error {
		if l.offset != nil {
			return errors.New("cannot open a line number and an offset")
		}
		l.line = &line
		return nil
	}
}

func Column(col int) Opt {
	return func(l *Location) error {
		if l.offset != nil {
			return errors.New("cannot open a column number and an offset")
		}
		l.col = &col
		return nil
	}
}

func Offset(offset int) Opt {
	return func(l *Location) error {
		if l.col != nil || l.line != nil {
			return errors.New("cannot open an offset and a line/column")
		}
		l.offset = &offset
		return nil
	}
}

func SkipUnbind() Opt {
	return func(l *Location) error {
		l.skipUnbind = true
		return nil
	}
}

func cp(ptr *int) *int {
	if ptr == nil {
		return nil
	}
	i := *ptr
	return &i
}

type Location struct {
	status.General
	driver gxui.Driver

	// These can only be set by FocusOpts in For().
	path              string
	offset, line, col *int
	skipUnbind        bool

	mover   Mover
	binder  Binder
	opener  EditorOpener
	openers []Opener

	binders  []FileBinder
	changers []FileChanger
}

func NewLocation(driver gxui.Driver) *Location {
	return &Location{driver: driver}
}

func (f *Location) For(opts ...Opt) bind.Bindable {
	newF := &Location{
		driver:     f.driver,
		path:       f.path,
		skipUnbind: f.skipUnbind,
		offset:     cp(f.offset),
		line:       cp(f.line),
		col:        cp(f.col),
	}
	newF.binders = append(newF.binders, f.binders...)
	newF.changers = append(newF.changers, f.changers...)
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
	return "focus-location"
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
		// TODO: scrap this and use hooks instead.
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
	if f.line == nil && f.col == nil && f.offset == nil {
		// The mover is not needed.
		return true
	}
	return f.mover != nil
}

func (f *Location) Exec() error {
	var oldPath string
	e := f.opener.CurrentEditor()
	if !f.skipUnbind && e != nil {
		oldPath = e.Filepath()
		f.binder.Pop()
	}
	path := f.path
	if path == "" && e != nil {
		path = e.Filepath()
	}
	if path == "" {
		return nil
	}
	e, _ = f.opener.Open(path)
	for _, o := range f.openers {
		o.Open(path)
	}
	if oldPath != path {
		for _, c := range f.changers {
			c.FileChanged(oldPath, path)
		}
	}
	var b []bind.Bindable
	for _, binder := range f.binders {
		b = append(b, binder.FileBindables(path)...)
	}
	f.binder.Push(b...)

	// Let the editor finish loading its text before we try
	// to load the start of a line.
	f.driver.Call(func() {
		f.moveCarets(e.(LineStarter))
	})
	return nil
}

func (f *Location) moveCarets(l LineStarter) {
	if f.offset == nil && f.line == nil && f.col == nil {
		return
	}
	if f.offset != nil {
		f.binder.Execute(f.mover.To(*f.offset))
		return
	}
	offset := 0
	if f.line != nil {
		offset = l.LineStart(*f.line)
	}
	if f.col != nil {
		offset += *f.col
	}
	f.binder.Execute(f.mover.To(offset))
}

func (f *Location) Bind(h bind.Bindable) (bind.HookedMultiOp, error) {
	newF := f.For().(*Location)
	switch src := h.(type) {
	case FileBinder:
		newF.binders = append(newF.binders, src)
	case FileChanger:
		newF.changers = append(newF.changers, src)
	default:
		return nil, fmt.Errorf("expected hook to be FileBinder or FileChanger, was %T", h)
	}
	return newF, nil
}
