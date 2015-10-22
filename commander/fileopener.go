// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package commander

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/gxui_playground/controller"
)

type FileOpener struct {
	file *FSLocator
	done bool
}

func NewFileOpener(driver gxui.Driver, theme *basic.Theme) controller.Command {
	fileOpener := new(FileOpener)
	fileOpener.Init(driver, theme)
	return fileOpener
}

func (f *FileOpener) Init(driver gxui.Driver, theme *basic.Theme) {
	f.file = NewFSLocator(driver, theme)
}

func (f *FileOpener) Name() string {
	return "open-file"
}

func (f *FileOpener) Start(control gxui.Control) gxui.Control {
	f.file.loadEditorDir(control)
	f.done = false
	return nil
}

func (f *FileOpener) Next() gxui.Focusable {
	if f.done {
		return nil
	}
	return f.file
}

func (f *FileOpener) Exec(element interface{}) (consume bool) {
	if editor, ok := element.(controller.Editor); ok {
		editor.Open(f.file.Path())
		return true
	}
	return false
}
