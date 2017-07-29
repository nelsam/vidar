// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package theme

type Color struct {
	R, G, B, A float32
}

type Highlight struct {
	Foreground, Background Color
}

type ConstructHighlights map[LanguageConstruct]Highlight

type Theme struct {
	Constructs ConstructHighlights

	// Rainbow is used to highlight any LanguageConstruct values
	// that are not in the Theme's Constructs map.  If you want
	// to guarantee rainbow parens in your syntax highlighting
	// plugin, use the ScopePair LanguageConstruct as your starting
	// value and increment for nested scopes.
	//
	// See the gosyntax plugin for an example.
	Rainbow *Rainbow
}
