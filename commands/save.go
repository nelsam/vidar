// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/editor"
)

type CurrentFileFinder interface {
	CurrentEditor() *editor.CodeEditor
	CurrentFile() string
}

type SaveCurrent struct {
	theme    gxui.Theme
	execErr  error
	filepath string
}

func NewSave(theme gxui.Theme) *SaveCurrent {
	return &SaveCurrent{theme: theme}
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
			s.execErr = fmt.Errorf("Could not stat file %s: %s", s.filepath, err)
			return true, false
		}
		if finfo.ModTime().After(editor.LastKnownMTime()) {
			// TODO: prompt for override
			s.execErr = fmt.Errorf("File %s changed on disk.  Cowardly refusing to overwrite.", s.filepath)
			return true, false
		}
	}
	f, err := os.Create(s.filepath)
	if err != nil {
		s.execErr = fmt.Errorf("Could not open %s for writing: %s", s.filepath, err)
		return true, false
	}
	defer f.Close()
	if !strings.HasSuffix(editor.Text(), "\n") {
		editor.SetText(editor.Text() + "\n")
	}
	if _, err := f.WriteString(editor.Text()); err != nil {
		s.execErr = fmt.Errorf("Could not write to file %s: %s", s.filepath, err)
		return true, false
	}
	s.execErr = nil
	editor.FlushedChanges()
	return true, true
}

func (s *SaveCurrent) statusInfo() (gxui.Color, string) {
	if s.execErr != nil {
		return commander.ColorErr, s.execErr.Error()
	}
	return commander.ColorInfo, fmt.Sprintf("Successfully saved %s", s.filepath)
}

func (s *SaveCurrent) Status() gxui.Control {
	label := s.theme.CreateLabel()
	color, message := s.statusInfo()
	label.SetColor(color)
	label.SetText(message)
	return label
}
