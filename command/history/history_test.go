// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package history_test

import (
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/command/history"
	"github.com/nelsam/vidar/command/input"
)

var (
	_ input.ChangeHook  = &history.History{}
	_ focus.FileChanger = &history.History{}
)
