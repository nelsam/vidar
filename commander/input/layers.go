// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package input

import "github.com/nelsam/vidar/theme"

type Span struct {
	Start, End int
}

type SyntaxLayer struct {
	Spans     []Span
	Construct theme.LanguageConstruct
}

func (l SyntaxLayer) Contains(pos int) bool {
	for _, s := range l.Spans {
		if s.Start <= pos && s.End >= pos {
			return true
		}
	}
	return false
}
