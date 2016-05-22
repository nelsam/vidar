// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/editor"
)

type CurrentFileFinder interface {
	CurrentEditor() *editor.CodeEditor
	CurrentFile() string
}

type SaveCurrent struct {
	statusKeeper

	filepath string
}

func NewSave(theme gxui.Theme) *SaveCurrent {
	return &SaveCurrent{
		statusKeeper: statusKeeper{theme: theme},
	}
}

func (s *SaveCurrent) Name() string {
	return "save-current-file"
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
	s.filepath = finder.CurrentFile()
	if !editor.LastKnownMTime().IsZero() {
		finfo, err := os.Stat(s.filepath)
		if err != nil {
			s.err = fmt.Sprintf("Could not stat file %s: %s", s.filepath, err)
			return true, false
		}
		if finfo.ModTime().After(editor.LastKnownMTime()) {
			// TODO: prompt for override
			s.err = fmt.Sprintf("File %s changed on disk.  Cowardly refusing to overwrite.", s.filepath)
			return true, false
		}
	}
	f, err := os.Create(s.filepath)
	if err != nil {
		s.err = fmt.Sprintf("Could not open %s for writing: %s", s.filepath, err)
		return true, false
	}
	defer f.Close()
	if !strings.HasSuffix(editor.Text(), "\n") {
		editor.SetText(editor.Text() + "\n")
	}
	if _, err := f.WriteString(editor.Text()); err != nil {
		s.err = fmt.Sprintf("Could not write to file %s: %s", s.filepath, err)
		return true, false
	}
	s.info = fmt.Sprintf("Successfully saved %s", s.filepath)
	editor.FlushedChanges()
	return true, true
}
