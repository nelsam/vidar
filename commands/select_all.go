// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

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

func (s *SelectAll) Exec(target interface{}) (executed, consume bool) {
	finder, ok := target.(EditorFinder)
	if !ok {
		return false, false
	}
	editor := finder.CurrentEditor()
	if editor == nil {
		return true, true
	}
	editor.SelectAll()
	return true, true
}
