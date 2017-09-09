// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import "github.com/nelsam/vidar/commander/input"

// Confirmer is a type that needs to know when a user is confirming
// an action.  It is typically a type that adds itself to the UI
// requesting that the user respond in some way.  The input handler
// will interpret user key presses and call all confirmers when a
// confirmation key (enter, in the default handler) is pressed.
type Confirmer interface {
	// Confirm will be called when a confirmation key is pressed.
	// It should return true if it should consume the confirmation
	// event.
	Confirm(input.Editor) (confirmed bool)
}
