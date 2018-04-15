// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package suggestion

import "fmt"

// A suggestion is a simple implementation of gxui.CodeSuggestion.
type Suggestion struct {
	Name      string
	Signature string
}

// String handles displaying the suggestion.
//
// TODO: Implement gxui.Viewer instead of gxui.Stringer, so that
//       we can syntax-highlight types in the completion list.
func (s Suggestion) String() string {
	return fmt.Sprintf("%s: %s", s.Name, s.Signature)
}
