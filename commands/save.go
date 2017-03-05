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
	"github.com/nelsam/vidar/settings"
)

type OpenProject interface {
	CurrentEditor() *editor.CodeEditor
	CurrentFile() string
	Project() settings.Project
}

type BeforeSaver interface {
	Name() string
	BeforeSave(proj settings.Project, path, contents string) (newContents string, err error)
}

type AfterSaver interface {
	Name() string
	AfterSave(proj settings.Project, path, contents string) error
}

type SaveCurrent struct {
	commander.GenericStatuser

	before []BeforeSaver
	after  []AfterSaver
}

func NewSave(theme gxui.Theme) *SaveCurrent {
	s := &SaveCurrent{}
	s.Theme = theme
	return s
}

func (s *SaveCurrent) Name() string {
	return "save-current-file"
}

func (s *SaveCurrent) Menu() string {
	return "File"
}

func (s *SaveCurrent) Clone() commander.CloneableCommand {
	newS := NewSave(s.Theme)
	newS.before = append(newS.before, s.before...)
	newS.after = append(newS.after, s.after...)
	return newS
}

func (s *SaveCurrent) Bind(h commander.CommandHook) error {
	switch src := h.(type) {
	case BeforeSaver:
		s.before = append(s.before, src)
	case AfterSaver:
		s.after = append(s.after, src)
	default:
		return fmt.Errorf("expected BeforeSaver or AfterSaver; got %T", h)
	}
	return nil
}

func (s *SaveCurrent) Exec(target interface{}) (executed, consume bool) {
	proj, ok := target.(OpenProject)
	if !ok {
		return false, false
	}
	editor := proj.CurrentEditor()
	if editor == nil {
		return true, true
	}

	filepath := proj.CurrentFile()
	if !editor.LastKnownMTime().IsZero() {
		finfo, err := os.Stat(filepath)
		if err != nil {
			s.Err = fmt.Sprintf("Could not stat file %s: %s", filepath, err)
			return true, false
		}
		if finfo.ModTime().After(editor.LastKnownMTime()) {
			// TODO: prompt for override
			s.Err = fmt.Sprintf("File %s changed on disk.  Cowardly refusing to overwrite.", filepath)
			return true, false
		}
	}

	text := editor.Text()
	formatted := text
	if !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}
	for _, b := range s.before {
		newText, err := b.BeforeSave(proj.Project(), filepath, text)
		if err != nil {
			s.Warn += fmt.Sprintf("%s: %s  ", b.Name(), err)
			continue
		}
		formatted = newText
	}

	if formatted != text {
		edits := []gxui.TextBoxEdit{
			{
				At:    0,
				Delta: len(formatted) - len(text),
				Old:   []rune(text),
				New:   []rune(formatted),
			},
		}
		editor.Controller().SetTextEdits([]rune(formatted), edits)
	}

	f, err := os.Create(filepath)
	if err != nil {
		s.Err = fmt.Sprintf("Could not open %s for writing: %s", filepath, err)
		return true, false
	}
	defer func() {
		f.Close()
		for _, a := range s.after {
			if err := a.AfterSave(proj.Project(), filepath, text); err != nil {
				s.Warn += fmt.Sprintf("%s: %s  ", a.Name(), err)
			}
		}
	}()

	if _, err := f.WriteString(text); err != nil {
		s.Err = fmt.Sprintf("Could not write to file %s: %s", filepath, err)
		return true, false
	}
	s.Info = fmt.Sprintf("Successfully saved %s", filepath)
	editor.FlushedChanges()
	return true, true
}
