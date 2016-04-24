package commands

import (
	"log"
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
	if editor == nil {
		return true, true
	}
	filepath := finder.CurrentFile()
	if !editor.LastKnownMTime().IsZero() {
		finfo, err := os.Stat(filepath)
		if err != nil {
			log.Printf("Error stating %s: %s", filepath, err)
			return true, false
		}
		if finfo.ModTime().After(editor.LastKnownMTime()) {
			// TODO: display an error, prompt for override
			log.Print("Error: Cowardly refusing to overwrite file (modified since last read)")
			return true, false
		}
	}
	f, err := os.Create(filepath)
	if err != nil {
		log.Printf("Failed to create %s: %s", filepath, err)
		return true, false
	}
	defer f.Close()
	if !strings.HasSuffix(editor.Text(), "\n") {
		editor.SetText(editor.Text() + "\n")
	}
	if _, err := f.WriteString(editor.Text()); err != nil {
		log.Printf("Error writing editor text to file %s: %s", filepath, err)
		return true, false
	}
	editor.FlushedChanges()
	return true, true
}
