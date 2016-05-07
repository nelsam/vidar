// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import "github.com/nelsam/vidar/editor"

type CurrentEditorCloser interface {
	CloseCurrentEditor() (string, *editor.CodeEditor)
}

type CloseTab struct{}

func NewCloseTab() *CloseTab {
	return &CloseTab{}
}

func (s *CloseTab) Name() string {
	return "close-current-tab"
}

func (s *CloseTab) Exec(target interface{}) (executed, consume bool) {
	closer, ok := target.(CurrentEditorCloser)
	if !ok {
		return false, false
	}
	closer.CloseCurrentEditor()
	return true, true
}
