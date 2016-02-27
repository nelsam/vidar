package commands

import (
	"os"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
)

type CurrentFileFinder interface {
	CurrentEditor() *editor.CodeEditor
	CurrentFile() string
}

type SaveCurrent struct{}

func NewSave() *SaveCurrent {
	return &SaveCurrent{}
}

func (s *SaveCurrent) Start(gxui.Control) gxui.Control {
	return nil
}

func (s *SaveCurrent) Name() string {
	return "save-current-file"
}

func (s *SaveCurrent) Next() gxui.Focusable {
	return nil
}

func (s *SaveCurrent) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(CurrentFileFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	filepath := finder.CurrentFile()
	if !editor.LastKnownMTime().IsZero() {
		finfo, err := os.Stat(filepath)
		if err != nil {
			panic(err)
		}
		if finfo.ModTime().After(editor.LastKnownMTime()) {
			// TODO: display an error, prompt for override
			panic("Cannot save file: written since last open")
		}
	}
	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if !strings.HasSuffix(editor.Text(), "\n") {
		editor.SetText(editor.Text() + "\n")
	}
	if _, err := f.WriteString(editor.Text()); err != nil {
		panic(err)
	}
	editor.FlushedChanges()
	return true, true
}
