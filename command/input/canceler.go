// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import "github.com/nelsam/vidar/commander/input"

// Canceler is a type that needs to know when it's being cancelled.
// This could be anything from a long-running process that can't
// make use of the context.Context (for whatever reason) to a type
// that adds itself to the UI and needs to know when it should be
// removed.
type Canceler interface {
	// Cancel is called when the hook should be cancelled.  It should
	// return true if it should consume the cancel event.
	Cancel(input.Editor) (cancelled bool)
}
