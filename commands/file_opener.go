// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"fmt"
	"go/token"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/editor"
)

type EditorOpener interface {
	Open(string, token.Position) (editor *editor.CodeEditor, existed bool)
}

type Opener interface {
	Open(string, token.Position)
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
	Pop()
}

type FileOpener struct {
	driver gxui.Driver
	theme  *basic.Theme

	file   *FSLocator
	cursor token.Position
	input  <-chan gxui.Focusable

	binder    Binder
	editor    *editor.CodeEditor
	skipHooks bool
	hooks     []FileBinder
}

func NewFileOpener(driver gxui.Driver, theme *basic.Theme) *FileOpener {
	fileOpener := new(FileOpener)
	fileOpener.Init(driver, theme)
	return fileOpener
}

func (f *FileOpener) Init(driver gxui.Driver, theme *basic.Theme) {
	f.driver = driver
	f.theme = theme
	f.file = NewFSLocator(driver, theme)
}

func (f *FileOpener) SetLocation(filepath string, position token.Position) {
	f.file.SetPath(filepath)
	f.cursor = position
}

func (f *FileOpener) Name() string {
	return "open-file"
}

func (f *FileOpener) Menu() string {
	return "File"
}

func (f *FileOpener) Start(control gxui.Control) gxui.Control {
	f.binder = nil
	f.editor = nil
	f.skipHooks = false

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

func (f *FileOpener) Exec(element interface{}) (executed, consume bool) {
	switch src := element.(type) {
	case EditorOpener:
		if f.editor != nil {
			return false, false
		}
		f.editor, f.skipHooks = src.Open(f.file.Path(), f.cursor)
		if f.binder != nil {
			f.setupHooks()
			return true, false
		}
	case Opener:
		src.Open(f.file.Path(), f.cursor)
		return false, false
	case Binder:
		if f.binder != nil {
			return false, false
		}
		f.binder = src
		if f.editor != nil {
			f.setupHooks()
			return true, false
		}
	}
	return false, false
}

func (f *FileOpener) setupHooks() {
	if f.skipHooks {
		return
	}
	path := f.file.Path()
	var b []bind.Bindable
	for _, h := range f.hooks {
		b = append(b, h.FileBindables(path)...)
	}
	binder := f.binder
	f.editor.OnGainedFocus(func() {
		binder.Push(b...)
	})
	f.editor.OnLostFocus(func() {
		binder.Pop()
	})
	if f.editor.HasFocus() {
		binder.Push(b...)
	}
}

func (f *FileOpener) Clone() commander.CloneableCommand {
	newF := NewFileOpener(f.driver, f.theme)
	newF.hooks = append(newF.hooks, f.hooks...)
	return newF
}

func (f *FileOpener) Bind(h bind.CommandHook) error {
	bndr, ok := h.(FileBinder)
	if !ok {
		return fmt.Errorf("expected hook to be FileBinder, was %T", h)
	}
	f.hooks = append(f.hooks, bndr)
	return nil
}
