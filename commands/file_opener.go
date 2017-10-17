// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/editor"
)

type EditorOpener interface {
	Open(string, int) (editor *editor.CodeEditor, existed bool)
	CurrentEditor() *editor.CodeEditor
}

type Opener interface {
	Open(string, int)
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
}

type FileOpener struct {
	driver gxui.Driver
	theme  *basic.Theme

	file   *FSLocator
	offset int
	input  <-chan gxui.Focusable

	binder  Binder
	opener  EditorOpener
	openers []Opener

	hooks []FileBinder
}

func NewFileOpener(driver gxui.Driver, theme *basic.Theme) *FileOpener {
	o := &FileOpener{
		driver: driver,
		theme:  theme,
		offset: -1,
	}
	o.file = NewFSLocator(driver, theme)
	return o
}

func (f *FileOpener) For(path string, offset int) bind.Bindable {
	newF := NewFileOpener(f.driver, f.theme)
	newF.hooks = append(newF.hooks, f.hooks...)

	newF.file.SetPath(path)
	newF.offset = offset
	return newF
}

func (f *FileOpener) Name() string {
	return "open-file"
}

func (f *FileOpener) Menu() string {
	return "File"
}

func (f *FileOpener) Start(control gxui.Control) gxui.Control {
	f.file.loadEditorDir(control)
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
	f.binder = nil
	f.opener = nil
	f.openers = nil
}

func (f *FileOpener) Store(elem interface{}) bind.Status {
	switch src := elem.(type) {
	case EditorOpener:
		f.opener = src
	case Opener:
		f.openers = append(f.openers, src)
	case Binder:
		f.binder = src
	}
	if f.opener != nil && f.binder != nil {
		return bind.Executing
	}
	return bind.Waiting
}

func (f *FileOpener) Exec() error {
	if f.opener.CurrentEditor() != nil {
		f.binder.Pop()
	}
	path := f.file.Path()
	f.opener.Open(path, f.offset)
	for _, o := range f.openers {
		o.Open(path, f.offset)
	}
	var b []bind.Bindable
	for _, h := range f.hooks {
		b = append(b, h.FileBindables(path)...)
	}
	f.binder.Push(b...)
	return nil
}

func (f *FileOpener) Bind(h bind.Bindable) (bind.HookedMultiOp, error) {
	bndr, ok := h.(FileBinder)
	if !ok {
		return nil, fmt.Errorf("expected hook to be FileBinder, was %T", h)
	}
	newF := NewFileOpener(f.driver, f.theme)
	newF.hooks = append(newF.hooks, f.hooks...)
	newF.hooks = append(newF.hooks, bndr)
	return newF, nil
}
