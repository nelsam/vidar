// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/settings"
)

type Projecter interface {
	Project() settings.Project
}

type Applier interface {
	Apply(input.Editor, ...input.Edit)
}

type Editor interface {
	input.Editor
	Filepath() string
	FlushedChanges()
	LastKnownMTime() time.Time
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
	status.General

	proj    *settings.Project
	applier Applier
	editor  Editor

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

func (s *SaveCurrent) Bind(h bind.CommandHook) error {
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

func (s *SaveCurrent) Reset() {
	s.proj = nil
	s.applier = nil
	s.editor = nil
}

func (s *SaveCurrent) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Projecter:
		proj := src.Project()
		s.proj = &proj
	case Applier:
		s.applier = src
	case Editor:
		s.editor = src
	}
	if s.editor != nil && s.proj != nil && s.applier != nil && s.editor != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (s *SaveCurrent) Exec() error {
	filepath := s.editor.Filepath()
	if !s.editor.LastKnownMTime().IsZero() {
		finfo, err := os.Stat(filepath)
		if err != nil {
			s.Err = fmt.Sprintf("Could not stat file %s: %s", filepath, err)
			return err
		}
		if finfo.ModTime().After(s.editor.LastKnownMTime()) {
			// TODO: prompt for override
			s.Err = fmt.Sprintf("File %s changed on disk.  Cowardly refusing to overwrite.", filepath)
			return err
		}
	}

	text := s.editor.Text()
	formatted := text
	if !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	proj := *s.proj
	for _, b := range s.before {
		newText, err := b.BeforeSave(proj, filepath, text)
		if err != nil {
			s.Warn += fmt.Sprintf("%s: %s  ", b.Name(), err)
			continue
		}
		formatted = newText
	}

	if formatted != text {
		s.applier.Apply(s.editor, input.Edit{
			At:  0,
			Old: []rune(text),
			New: []rune(formatted),
		})
		text = formatted
	}

	f, err := os.Create(filepath)
	if err != nil {
		s.Err = fmt.Sprintf("Could not open %s for writing: %s", filepath, err)
		return err
	}
	defer func() {
		f.Close()
		for _, a := range s.after {
			if err := a.AfterSave(proj, filepath, text); err != nil {
				s.Warn += fmt.Sprintf("%s: %s  ", a.Name(), err)
			}
		}
	}()

	if _, err := f.WriteString(text); err != nil {
		s.Err = fmt.Sprintf("Could not write to file %s: %s", filepath, err)
		return err
	}
	s.Info = fmt.Sprintf("Successfully saved %s", filepath)
	s.editor.FlushedChanges()
	return nil
}
