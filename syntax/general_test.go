// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/nelsam/vidar/syntax"
)

func TestUnicode(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

func µ() string {
	var þ = "Ωð"
	return þ
}
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(4)
	expectPositionMatch(t, src, "var", keywords.Spans()[2])
	expectPositionMatch(t, src, "return", keywords.Spans()[3])

	strings := layers[syntax.DefaultTheme.Colors.String]
	expect(strings.Spans()).To.Have.Len(1)
	expectPositionMatch(t, src, `"Ωð"`, strings.Spans()[0])
}

func TestPackageDocs(t *testing.T) {
	expect := expect.New(t)

	src := `
// Package foo does stuff.
// It is also a thing.
package foo
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(src)
	expect(err).To.Be.Nil()
	layers := s.Layers()
	expect(layers).To.Have.Len(2)

	comments := layers[syntax.DefaultTheme.Colors.Comment]
	expect(comments.Spans()).To.Have.Len(1)
	comment := "// Package foo does stuff.\n" +
		"// It is also a thing."
	expectPositionMatch(t, src, comment, comments.Spans()[0])
}
