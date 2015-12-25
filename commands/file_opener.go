// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commands

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
)

type opener interface {
	Open(string, int)
}

type FileOpener struct {
	file   *FSLocator
	cursor int
	input  <-chan gxui.Focusable
}

func NewFileOpener(driver gxui.Driver, theme *basic.Theme) *FileOpener {
	fileOpener := new(FileOpener)
	fileOpener.Init(driver, theme)
	return fileOpener
}

func (f *FileOpener) Init(driver gxui.Driver, theme *basic.Theme) {
	f.file = NewFSLocator(driver, theme)
}

func (f *FileOpener) SetLocation(filepath string, position int) {
	f.file.SetPath(filepath)
	f.cursor = position
}

func (f *FileOpener) Name() string {
	return "open-file"
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

func (f *FileOpener) Exec(element interface{}) (executed, consume bool) {
	if opener, ok := element.(opener); ok {
		opener.Open(f.file.Path(), f.cursor)
		return true, false
	}
	return false, false
}
