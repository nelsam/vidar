// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"github.com/nelsam/gxui"
)

type AllSaver interface {
	SaveAll()
}

type SaveAll struct{}

func NewSaveAll(theme gxui.Theme) *SaveAll {
	return &SaveAll{}
}

func (s *SaveAll) Name() string {
	return "save-all-files"
}

func (s *SaveAll) Menu() string {
	return "File"
}

func (s *SaveAll) Exec(target interface{}) (executed, consume bool) {
	saver, ok := target.(AllSaver)
	if !ok {
		return false, false
	}
	saver.SaveAll()
	return true, true
}
