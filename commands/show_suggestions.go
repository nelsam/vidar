// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

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

func (s *ShowSuggestions) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	finder.CurrentEditor().ShowSuggestionList()
	return true, true
}
