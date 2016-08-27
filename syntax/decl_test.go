// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"strings"
	"testing"

	"github.com/a8m/expect"
	"github.com/nelsam/vidar/syntax"
)

func TestDecl(t *testing.T) {
	t.Run("Gen", Gen)
	t.Run("Func", Func)
	t.Run("Bad", Bad)
}

func Gen(t *testing.T) {
	t.Run("NoParens", NoParen)
	t.Run("Parens", Paren)
}

func NoParen(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

// Foo is a thing
var Foo string
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(3)

	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(2)
	expectPositionMatch(t, src, "var", keywords.Spans()[1])

	comments := layers[syntax.DefaultTheme.Colors.Comment]
	expect(comments.Spans()).To.Have.Len(1)
	expectPositionMatch(t, src, "// Foo is a thing", comments.Spans()[0])

	typs := layers[syntax.DefaultTheme.Colors.Type]
	expect(typs.Spans()).To.Have.Len(1)
	expectPositionMatch(t, src, "string", typs.Spans()[0])
}

func Paren(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

var (
	Foo string
	Bar int
)
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(3)

	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(2)
	expectPositionMatch(t, src, "var", keywords.Spans()[1])

	parens := layers[syntax.DefaultTheme.Rainbow.New()]
	syntax.DefaultTheme.Rainbow.Pop()
	expect(parens.Spans()).To.Have.Len(2)
	expectPositionMatch(t, src, "(", parens.Spans()[0])
	expectPositionMatch(t, src, ")", parens.Spans()[1])

	typs := layers[syntax.DefaultTheme.Colors.Type]
	expect(typs.Spans()).To.Have.Len(2)
	expectPositionMatch(t, src, "string", typs.Spans()[0])
	expectPositionMatch(t, src, "int", typs.Spans()[1])
}

func Func(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

func Foo(bar string) int {
	return 0
}
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(5)

	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(3)
	expectPositionMatch(t, src, "func", keywords.Spans()[1])
	expectPositionMatch(t, src, "return", keywords.Spans()[2])

	parens := layers[syntax.DefaultTheme.Rainbow.New()]
	syntax.DefaultTheme.Rainbow.Pop()
	expect(parens.Spans()).To.Have.Len(4)
	expectPositionMatch(t, src, "(", parens.Spans()[0])
	expectPositionMatch(t, src, ")", parens.Spans()[1])
	expectPositionMatch(t, src, "{", parens.Spans()[2])
	expectPositionMatch(t, src, "}", parens.Spans()[3])

	typs := layers[syntax.DefaultTheme.Colors.Type]
	expect(typs.Spans()).To.Have.Len(2)
	expectPositionMatch(t, src, "string", typs.Spans()[0])
	expectPositionMatch(t, src, "int", typs.Spans()[1])

	ints := layers[syntax.DefaultTheme.Colors.Num]
	expect(ints.Spans()).To.Have.Len(1)
	expectPositionMatch(t, src, "0", ints.Spans()[0])
}

func Bad(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

10
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(src)
	expect(err).Not.To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(2)

	bad := layers[syntax.DefaultTheme.Colors.Bad]
	expect(bad.Spans()).To.Have.Len(1)
	expectedStart := strings.Index(src, "10")
	start, _ := bad.Spans()[0].Range()
	expect(start).To.Equal(expectedStart)
}
