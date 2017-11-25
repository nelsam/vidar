// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"strings"
	"testing"

	"github.com/a8m/expect"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"
)

func TestDecl(t *testing.T) {
	t.Run("Gen", GenDecl)
	t.Run("Func", FuncDecl)
	t.Run("Bad", BadDecl)
}

func GenDecl(t *testing.T) {
	t.Run("NoParens", GenDeclNoParen)
	t.Run("Parens", GenDeclParen)
}

func GenDeclNoParen(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

// Foo is a thing
var Foo string
`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(3)

	keywords := layers[theme.Keyword]
	expect(keywords.Spans).To.Have.Len(2)
	expect(keywords.Spans[1]).To.Pass(position{src: src, match: "var"})

	comments := layers[theme.Comment]
	expect(comments.Spans).To.Have.Len(1)
	expect(comments.Spans[0]).To.Pass(position{src: src, match: "// Foo is a thing"})

	typs := layers[theme.Type]
	expect(typs.Spans).To.Have.Len(1)
	expect(typs.Spans[0]).To.Pass(position{src: src, match: "string"})
}

func GenDeclParen(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

var (
	Foo string
	Bar int
)
`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(3)

	keywords := layers[theme.Keyword]
	expect(keywords.Spans).To.Have.Len(2)
	expect(keywords.Spans[1]).To.Pass(position{src: src, match: "var"})

	parens := layers[theme.ScopePair]
	expect(parens.Spans).To.Have.Len(2)
	expect(parens.Spans[0]).To.Pass(position{src: src, match: "("})
	expect(parens.Spans[1]).To.Pass(position{src: src, match: ")"})

	typs := layers[theme.Type]
	expect(typs.Spans).To.Have.Len(2)
	expect(typs.Spans[0]).To.Pass(position{src: src, match: "string"})
	expect(typs.Spans[1]).To.Pass(position{src: src, match: "int"})
}

func FuncDecl(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

func Foo(bar string) int {
	return 0
}
`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(5)

	keywords := layers[theme.Keyword]
	expect(keywords.Spans).To.Have.Len(3)
	expect(keywords.Spans[1]).To.Pass(position{src: src, match: "func"})
	expect(keywords.Spans[2]).To.Pass(position{src: src, match: "return"})

	parens := layers[theme.ScopePair]
	expect(parens.Spans).To.Have.Len(4)
	expect(parens.Spans[0]).To.Pass(position{src: src, match: "("})
	expect(parens.Spans[1]).To.Pass(position{src: src, match: ")"})
	expect(parens.Spans[2]).To.Pass(position{src: src, match: "{"})
	expect(parens.Spans[3]).To.Pass(position{src: src, match: "}"})

	typs := layers[theme.Type]
	expect(typs.Spans).To.Have.Len(2)
	expect(typs.Spans[0]).To.Pass(position{src: src, match: "string"})
	expect(typs.Spans[1]).To.Pass(position{src: src, match: "int"})

	ints := layers[theme.Num]
	expect(ints.Spans).To.Have.Len(1)
	expect(ints.Spans[0]).To.Pass(position{src: src, match: "0"})
}

func BadDecl(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo

10
`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).Not.To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(2)

	bad := layers[theme.Bad]
	expect(bad.Spans).To.Have.Len(1)
	expectedStart := strings.Index(src, "10")
	expect(bad.Spans[0].Start).To.Equal(expectedStart)
}
