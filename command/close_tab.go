// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/input"
)

type CurrentEditorCloser interface {
	CloseCurrentEditor() (string, input.Editor)
	CurrentEditor() input.Editor
}

type BindPopper interface {
	Pop() []bind.Bindable
}

type CloseTab struct {
	closer CurrentEditorCloser
	binder BindPopper
}

func NewCloseTab() *CloseTab {
	return &CloseTab{}
}

func (s *CloseTab) Name() string {
	return "close-current-tab"
}

func (s *CloseTab) Menu() string {
	return "File"
}

func (s *CloseTab) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl,
		Key:      gxui.KeyW,
	}}
}

func (s *CloseTab) Reset() {
	s.closer = nil
	s.binder = nil
}

func (s *CloseTab) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case CurrentEditorCloser:
		s.closer = src
	case BindPopper:
		s.binder = src
	}
	if s.closer != nil && s.binder != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (s *CloseTab) Exec() error {
	s.closer.CloseCurrentEditor()
	if s.closer.CurrentEditor() == nil {
		s.binder.Pop()
	}
	return nil
}
