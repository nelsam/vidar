// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import "github.com/nelsam/vidar/commander/bind"

type ShowSuggestions struct {
}

func NewShowSuggestions() *ShowSuggestions {
	return &ShowSuggestions{}
}

func (s *ShowSuggestions) Name() string {
	return "show-suggestions"
}

func (s *ShowSuggestions) Menu() string {
	return "Edit"
}

func (s *ShowSuggestions) Exec(target interface{}) bind.Status {
	finder, ok := target.(EditorFinder)
	if !ok {
		return bind.Waiting
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return bind.Done
	}
	editor.ShowSuggestionList()
	return bind.Done
}
