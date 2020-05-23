// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package text

// Editor is the local set of methods that the Editor type passed
// to InputHandlers are guaranteed to have.  Other methods may be
// accessed with type assertions, but we pass the Editor type
// instead of *editor.Editor in order to allow the more common
// methods to be accessed without needing to import the editor
// package.
//
// Due to the oddities in dependency versions and plugins, this
// helps plugins avoid needing to be rebuilt every time the editor
// package changes.
type Editor interface {
	Filepath() string
	Text() string
	Runes() []rune
	SetText(string)
	SyntaxLayers() []SyntaxLayer
	SetSyntaxLayers([]SyntaxLayer)
}
