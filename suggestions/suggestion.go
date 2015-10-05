// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package suggestions

import "fmt"

// A suggestion is a simple implementation of gxui.CodeSuggestion.
type suggestion struct {
	Value, Type string
}

// String handles displaying the suggestion.
//
// TODO: Implement gxui.Viewer instead of gxui.Stringer, so that
//       we can syntax-highlight types in the completion list.
func (c suggestion) String() string {
	return fmt.Sprintf("%s: %s", c.Value, c.Type)
}

func (c suggestion) Name() string {
	return c.Value
}

func (c suggestion) Code() string {
	return c.Value
}
