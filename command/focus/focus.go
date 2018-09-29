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

// EditorOpener represents a type that can open a file, returning
// an editor.
type EditorOpener interface {
	Open(string) (editor input.Editor, existed bool)
	CurrentEditor() input.Editor
}

// Opener represents a type that can open a file, but doesn't
// return anything.
//
// This is legacy code from before hooks existed and should be
// removed when all places that currently use it are converted to
// using hooks.
type Opener interface {
	Open(string)
}

// LineStarter represents a type that knows which position lines start
// at.
type LineStarter interface {
	LineStart(int) int
}

// Mover represents a type that can move carets.
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

// Opt is an option function that will apply to *Location types.
type Opt func(*Location) error

// Path takes a file path and returns an Opt that will modify a
// *Location to focus that file.
func Path(path string) Opt {
	return func(l *Location) error {
		l.path = path
		return nil
	}
}

// Line takes a line number and returns an Opt that will modify a
// *Location to move carets to that line.  Compatible with Column.
func Line(line int) Opt {
	return func(l *Location) error {
		if l.offset != nil {
			return errors.New("cannot open a line number and an offset")
		}
		l.line = &line
		return nil
	}
}

// Column takes a column number and returns an Opt that will modify
// a *Location to move carets to that column.  Compatible with Line.
func Column(col int) Opt {
	return func(l *Location) error {
		if l.offset != nil {
			return errors.New("cannot open a column number and an offset")
		}
		l.col = &col
		return nil
	}
}

// Offset takes a character offset and returns an Opt that will
// modify a *Location to move carets to that offset.
func Offset(offset int) Opt {
	return func(l *Location) error {
		if l.col != nil || l.line != nil {
			return errors.New("cannot open an offset and a line/column")
		}
		l.offset = &offset
		return nil
	}
}

// SkipUnbind returns an Opt that modifies a *Location to skip the
// process of unbinding previously-bound commands.  This is currently
// used to work around an issue that would leave the editor with *no*
// commands bound to it, as the final entry got popped ofl.
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

// Location is a bind.MultiOp that knows how to focus locations in
// the code.
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

// NewLocation returns a *Location bound to the passed in driver.
func NewLocation(driver gxui.Driver) *Location {
	return &Location{driver: driver}
}

// For makes a clone of l, then applies opts to the copy before
// returning it.
func (l *Location) For(opts ...Opt) bind.Bindable {
	newL := &Location{
		driver:     l.driver,
		path:       l.path,
		skipUnbind: l.skipUnbind,
		offset:     cp(l.offset),
		line:       cp(l.line),
		col:        cp(l.col),
	}
	newL.binders = append(newL.binders, l.binders...)
	newL.changers = append(newL.changers, l.changers...)
	for _, o := range opts {
		if err := o(newL); err != nil {
			if len(newL.Warn) != 0 {
				newL.Warn += "; "
			}
			newL.Warn += fmt.Sprintf("failed to apply option: %s", err)
		}
	}
	return newL
}

// Name returns l's name.
func (l *Location) Name() string {
	return "focus-location"
}

// Reset resets l's execution state.
func (l *Location) Reset() {
	l.mover = nil
	l.binder = nil
	l.opener = nil
	l.openers = nil
}

// Store checks elem for an types that l needs to store in order to
// execute.
func (l *Location) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case Mover:
		l.mover = src
	case EditorOpener:
		l.opener = src
	case Opener:
		// TODO: scrap this and use hooks instead.
		l.openers = append(l.openers, src)
	case Binder:
		l.binder = src
	}

	if !l.moverReady() || l.opener == nil || l.binder == nil {
		return bind.Waiting
	}
	return bind.Executing
}

func (l *Location) moverReady() bool {
	if l.line == nil && l.col == nil && l.offset == nil {
		// The mover is not needed.
		return true
	}
	return l.mover != nil
}

// Exec executes l against the values it has stored, returning an
// error if it encounters any problems.
func (l *Location) Exec() error {
	var oldPath string
	e := l.opener.CurrentEditor()
	if !l.skipUnbind && e != nil {
		oldPath = e.Filepath()
		l.binder.Pop()
	}
	path := l.path
	if path == "" && e != nil {
		path = e.Filepath()
	}
	if path == "" {
		return nil
	}
	e, _ = l.opener.Open(path)
	for _, o := range l.openers {
		o.Open(path)
	}
	if oldPath != path {
		for _, c := range l.changers {
			c.FileChanged(oldPath, path)
		}
	}
	var b []bind.Bindable
	for _, binder := range l.binders {
		b = append(b, binder.FileBindables(path)...)
	}
	l.binder.Push(b...)

	// Let the editor finish loading its text before we try
	// to load the start of a line.
	l.driver.Call(func() {
		l.moveCarets(e.(LineStarter))
	})
	return nil
}

func (l *Location) moveCarets(s LineStarter) {
	if l.offset == nil && l.line == nil && l.col == nil {
		return
	}
	if l.offset != nil {
		l.binder.Execute(l.mover.To(*l.offset))
		return
	}
	offset := 0
	if l.line != nil {
		offset = s.LineStart(*l.line)
	}
	if l.col != nil {
		offset += *l.col
	}
	l.binder.Execute(l.mover.To(offset))
}

// Bind binds hooks to l.
func (l *Location) Bind(h bind.Bindable) (bind.HookedMultiOp, error) {
	newF := l.For().(*Location)
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
