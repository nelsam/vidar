// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

type Applier interface {
	Apply(text.Editor, ...text.Edit)
}

type SaveEditor interface {
	text.Editor
	FlushedChanges()
	LastKnownMTime() time.Time
}

type Projecter interface {
	Project() setting.Project
}

type BeforeSaver interface {
	Name() string
	BeforeSave(proj setting.Project, path, contents string) (newContents string, err error)
}

type AfterSaver interface {
	Name() string
	AfterSave(proj setting.Project, path, contents string) error
}

type SaveCurrent struct {
	status.General

	proj    *setting.Project
	applier Applier
	editor  SaveEditor

	before []BeforeSaver
	after  []AfterSaver
}

type LabelCreator interface {
	CreateLabel() gxui.Label
}

func NewSave(theme LabelCreator) *SaveCurrent {
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

func (s *SaveCurrent) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyS,
	}}
}

func (s *SaveCurrent) Bind(h bind.Bindable) (bind.HookedMultiOp, error) {
	newS := NewSave(s.Theme)
	newS.before = append(newS.before, s.before...)
	newS.after = append(newS.after, s.after...)
	switch src := h.(type) {
	case BeforeSaver:
		newS.before = append(newS.before, src)
	case AfterSaver:
		newS.after = append(newS.after, src)
	default:
		return nil, fmt.Errorf("expected BeforeSaver or AfterSaver; got %T", h)
	}
	return newS, nil
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
	case SaveEditor:
		s.editor = src
	}
	if s.editor != nil && s.proj != nil && s.applier != nil {
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

	current := s.editor.Text()
	formatted := current
	if !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	proj := *s.proj
	for _, b := range s.before {
		newText, err := b.BeforeSave(proj, filepath, current)
		if err != nil {
			s.Warn += fmt.Sprintf("%s: %s  ", b.Name(), err)
			continue
		}
		formatted = newText
	}

	if formatted != current {
		s.applier.Apply(s.editor, text.Edit{
			At:  0,
			Old: []rune(current),
			New: []rune(formatted),
		})
		current = formatted
	}

	f, err := os.Create(filepath)
	if err != nil {
		s.Err = fmt.Sprintf("Could not open %s for writing: %s", filepath, err)
		return err
	}
	defer func() {
		f.Close()
		for _, a := range s.after {
			if err := a.AfterSave(proj, filepath, current); err != nil {
				s.Warn += fmt.Sprintf("%s: %s  ", a.Name(), err)
			}
		}
	}()

	if _, err := f.WriteString(current); err != nil {
		s.Err = fmt.Sprintf("Could not write to file %s: %s", filepath, err)
		return err
	}
	s.Info = fmt.Sprintf("Successfully saved %s", filepath)
	s.editor.FlushedChanges()
	return nil
}
