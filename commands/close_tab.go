// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICloseTabCENCloseTabSE file.

package commands

import "github.com/nelsam/gxui"

type CurrentTabCloser interface {
	CloseCurrentTab()
}

type CloseTab struct{}

func NewCloseTab() *CloseTab {
	return &CloseTab{}
}

func (s *CloseTab) Start(gxui.Control) gxui.Control {
	return nil
}

func (s *CloseTab) Name() string {
	return "close-current-tab"
}

func (s *CloseTab) Next() gxui.Focusable {
	return nil
}

func (s *CloseTab) Exec(target interface{}) (executed, consume bool) {
	closer, ok := target.(CurrentTabCloser)
	if !ok {
		return false, false
	}
	closer.CloseCurrentTab()
	return true, true
}
