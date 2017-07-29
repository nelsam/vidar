// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import "github.com/nelsam/vidar/commander/bind"

type SelectAll struct{}

func NewSelectAll() *SelectAll {
	return &SelectAll{}
}

func (s *SelectAll) Name() string {
	return "select-all"
}

func (s *SelectAll) Menu() string {
	return "File"
}

func (s *SelectAll) Exec(target interface{}) bind.Status {
	finder, ok := target.(EditorFinder)
	if !ok {
		return bind.Waiting
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return bind.Done
	}
	editor.SelectAll()
	return bind.Done
}
