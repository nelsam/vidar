// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import "github.com/nelsam/vidar/commander/text"

// An AppliedChangeHook is a hook that needs to be informed about
// edits immediately after they are applied, before the call to
// Apply is finished.  Its Applied method will be called
// synchronously in the Handler, so it should be as minimal as
// possible.
type AppliedChangeHook interface {
	Applied(text.Editor, []text.Edit)
}
