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

func GenDecl(t *testing.T) {
	t.Run("NoParens", NoParen)
	t.Run("Parens", Paren)
}

func NoParen(t *testing.T) {
	expect := expect.New(t)

	ast := `
package foo

// Foo is a thing
var Foo string
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(ast)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(3)

	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(2)

	// var keyword
	start, end := keywords.Spans()[1].Range()
	expectedStart := strings.Index(ast, "var")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("var"))

	comments := layers[syntax.DefaultTheme.Colors.Comment]
	expect(comments.Spans()).To.Have.Len(1)
	start, end = comments.Spans()[0].Range()
	expectedStart = strings.Index(ast, "//")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("// Foo is a thing"))

	typs := layers[syntax.DefaultTheme.Colors.Type]
	expect(typs.Spans()).To.Have.Len(1)
	start, end = typs.Spans()[0].Range()
	expectedStart = strings.Index(ast, "string")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("string"))
}

func Paren(t *testing.T) {
	expect := expect.New(t)

	ast := `
package foo

var (
	Foo string
	Bar int
)
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(ast)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(3)

	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(2)

	// var keyword
	start, end := keywords.Spans()[1].Range()
	expectedStart := strings.Index(ast, "var")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("var"))

	typs := layers[syntax.DefaultTheme.Colors.Type]
	expect(typs.Spans()).To.Have.Len(2)
	start, end = typs.Spans()[0].Range()
	expectedStart = strings.Index(ast, "string")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("string"))
	start, end = typs.Spans()[1].Range()
	expectedStart = strings.Index(ast, "int")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("int"))
}

func FuncDecl(t *testing.T) {
	expect := expect.New(t)

	ast := `
package foo

func Foo(bar string) int {
	return 0
}
`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(ast)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(5)
	// TODO: test paren colors from rainbow parens

	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(3)

	// func keyword
	start, end := keywords.Spans()[1].Range()
	expectedStart := strings.Index(ast, "func")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("func"))

	// return keyword
	start, end = keywords.Spans()[2].Range()
	expectedStart = strings.Index(ast, "return")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("return"))

	typs := layers[syntax.DefaultTheme.Colors.Type]
	expect(typs.Spans()).To.Have.Len(2)
	start, end = typs.Spans()[0].Range()
	expectedStart = strings.Index(ast, "string")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("string"))
	start, end = typs.Spans()[1].Range()
	expectedStart = strings.Index(ast, "int")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + len("int"))

	ints := layers[syntax.DefaultTheme.Colors.Num]
	expect(ints.Spans()).To.Have.Len(1)
	start, end = ints.Spans()[0].Range()
	expectedStart = strings.Index(ast, "0")
	expect(start).To.Equal(expectedStart)
	expect(end).To.Equal(expectedStart + 1)
}
