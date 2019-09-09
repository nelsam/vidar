// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history_test

import (
	"github.com/nelsam/vidar/command/history"
	"github.com/nelsam/vidar/commander/bind"
)

type T interface {
	Helper()
	Fatal(...interface{})
}

func findHistory(t T, all []bind.Bindable) *history.History {
	t.Helper()
	for _, b := range all {
		if h, ok := b.(*history.History); ok {
			return h
		}
	}
	t.Fatal("could not find *history.History type")
	return nil
}
