// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"strings"
	"testing"

	"github.com/apoydence/onpar"
	"github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"
)

func TestDecl(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) expect.Expectation {
		return expect.New(t)
	})

	o.Group("Gen", func() {
		o.Group("NoParen", func() {
			const src = `
			package foo
			
			// Foo is a thing
			var Foo string
			`

			o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
				s := syntax.New()
				err := s.Parse(src)
				expect(err).To(BeNil())

				layers := s.Layers()
				expect(layers).To(HaveLen(3))
				return expect, layers
			})

			o.Spec("it highlights keywords", func(expect expect.Expectation, layers []input.SyntaxLayer) {
				keywords := findLayer(theme.Keyword, layers)
				expect(keywords.Spans).To(HaveLen(2))
				expect(keywords.Spans[1]).To(matchPosition{src: src, match: "var"})
			})

			o.Spec("it highlights comments", func(expect expect.Expectation, layers []input.SyntaxLayer) {
				comments := findLayer(theme.Comment, layers)
				expect(comments.Spans).To(HaveLen(1))
				expect(comments.Spans[0]).To(matchPosition{src: src, match: "// Foo is a thing"})
			})

			o.Spec("it highlights types", func(expect expect.Expectation, layers []input.SyntaxLayer) {
				typs := findLayer(theme.Type, layers)
				expect(typs.Spans).To(HaveLen(1))
				expect(typs.Spans[0]).To(matchPosition{src: src, match: "string"})
			})
		})

		o.Group("Paren", func() {

			const src = `
			package foo
			
			var (
				Foo string
				Bar int
			)
			`

			o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
				s := syntax.New()
				err := s.Parse(src)
				expect(err).To(BeNil())

				layers := s.Layers()
				expect(layers).To(HaveLen(3))

				return expect, layers
			})

			o.Spec("it highlights keywords", func(expect expect.Expectation, layers []input.SyntaxLayer) {
				keywords := findLayer(theme.Keyword, layers)
				expect(keywords.Spans).To(HaveLen(2))
				expect(keywords.Spans[1]).To(matchPosition{src: src, match: "var"})
			})

			o.Spec("it highlights parenthesis", func(expect expect.Expectation, layers []input.SyntaxLayer) {
				parens := findLayer(theme.ScopePair, layers)
				expect(parens.Spans).To(HaveLen(2))
				expect(parens.Spans[0]).To(matchPosition{src: src, match: "("})
				expect(parens.Spans[1]).To(matchPosition{src: src, match: ")"})
			})

			o.Spec("it highlights types", func(expect expect.Expectation, layers []input.SyntaxLayer) {
				typs := findLayer(theme.Type, layers)
				expect(typs.Spans).To(HaveLen(2))
				expect(typs.Spans[0]).To(matchPosition{src: src, match: "string"})
				expect(typs.Spans[1]).To(matchPosition{src: src, match: "int"})
			})
		})
	})

	o.Group("Func", func() {
		// Note: this needs a trailing newline to parse the closing brace correctly.
		const src = `
		package foo
			
		func Foo(bar string) int {
			return 0
		}
		`

		o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
			s := syntax.New()
			err := s.Parse(src)
			expect(err).To(BeNil())

			layers := s.Layers()
			expect(layers).To(HaveLen(5))

			return expect, layers
		})

		o.Spec("it highlights function and return keywords", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			keywords := findLayer(theme.Keyword, layers)
			expect(keywords.Spans).To(HaveLen(3))
			expect(keywords.Spans[1]).To(matchPosition{src: src, match: "func"})
			expect(keywords.Spans[2]).To(matchPosition{src: src, match: "return"})
		})

		o.Spec("it highlights function parenthesis and braces", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			parens := findLayer(theme.ScopePair, layers)
			expect(parens.Spans).To(HaveLen(4))
			expect(parens.Spans[0]).To(matchPosition{src: src, match: "("})
			expect(parens.Spans[1]).To(matchPosition{src: src, match: ")"})
			expect(parens.Spans[2]).To(matchPosition{src: src, match: "{"})
			expect(parens.Spans[3]).To(matchPosition{src: src, match: "}"})
		})

		o.Spec("it highlights parameter and return types", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			typs := findLayer(theme.Type, layers)
			expect(typs.Spans).To(HaveLen(2))
			expect(typs.Spans[0]).To(matchPosition{src: src, match: "string"})
			expect(typs.Spans[1]).To(matchPosition{src: src, match: "int"})
		})

		o.Spec("it highlights expressions in the return statement", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			ints := findLayer(theme.Num, layers)
			expect(ints.Spans).To(HaveLen(1))
			expect(ints.Spans[0]).To(matchPosition{src: src, match: "0"})
		})
	})

	o.Group("Bad", func() {
		const src = `
		package foo
		
		10`

		o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
			s := syntax.New()
			err := s.Parse(src)
			expect(err).To(Not(BeNil()))

			layers := s.Layers()
			expect(layers).To(HaveLen(2))

			return expect, layers
		})

		o.Spec("it highlights bad syntax", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			bad := findLayer(theme.Bad, layers)
			expect(bad.Spans).To(HaveLen(1))
			expectedStart := strings.Index(src, "10")
			expect(bad.Spans[0].Start).To(Equal(expectedStart))
		})
	})
}
