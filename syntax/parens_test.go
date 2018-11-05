// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"testing"

	"github.com/apoydence/onpar"
	"github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"
)

func TestParens(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) expect.Expectation {
		return expect.New(t)
	})

	o.Group("FunctionBody", func() {
		// Note: there needs to be a newline after the closing brace for it to
		// parse correctly.
		const src = `
		package foo
	
		func (*Foo) Foo(baz Baz) {
			baz.Bar()
		}
		`

		o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
			s := syntax.New()
			err := s.Parse(src)
			expect(err).To(BeNil())

			return expect, s.Layers()
		})

		o.Spec("it highlights the outer-most scope pairs with one color", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			outerScope := findLayer(theme.ScopePair, layers)
			expect(outerScope.Spans).To(HaveLen(6))

			expect(outerScope.Spans[0]).To(matchPosition{src: src, match: "("})
			expect(outerScope.Spans[1]).To(matchPosition{src: src, match: ")"})
			expect(outerScope.Spans[2]).To(matchPosition{src: src, match: "(", idx: 1})
			expect(outerScope.Spans[3]).To(matchPosition{src: src, match: ")", idx: 1})
			expect(outerScope.Spans[4]).To(matchPosition{src: src, match: "{"})
			expect(outerScope.Spans[5]).To(matchPosition{src: src, match: "}"})
		})

		o.Spec("it highlights the nested scope pairs with a different color", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			secondParens := findLayer(theme.ScopePair+1, layers)
			expect(secondParens.Spans).To(HaveLen(2))
			expect(secondParens.Spans[0]).To(matchPosition{src: src, match: "(", idx: 2})
			expect(secondParens.Spans[1]).To(matchPosition{src: src, match: ")", idx: 2})
		})
	})

	o.Group("Array", func() {
		const src = `
		package foo
		
		func main() {
			var v [3]string
			v := [...]string{"foo"}
		}`

		o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
			s := syntax.New()
			err := s.Parse(src)
			expect(err).To(BeNil())

			return expect, s.Layers()
		})

		o.Spec("it highlights array brackets and braces in a nested scope", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			arrParens := findLayer(theme.ScopePair+1, layers)
			expect(arrParens.Spans).To(HaveLen(6))
			expect(arrParens.Spans[0]).To(matchPosition{src: src, match: "["})
			expect(arrParens.Spans[1]).To(matchPosition{src: src, match: "]"})
			expect(arrParens.Spans[2]).To(matchPosition{src: src, match: "[", idx: 1})
			expect(arrParens.Spans[3]).To(matchPosition{src: src, match: "]", idx: 1})
			expect(arrParens.Spans[4]).To(matchPosition{src: src, match: "{", idx: 1})
			expect(arrParens.Spans[5]).To(matchPosition{src: src, match: "}"})
		})

		o.Spec("it highlights array types", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			typs := findLayer(theme.Type, layers)
			expect(typs.Spans).To(HaveLen(2))
			expect(typs.Spans[0]).To(matchPosition{src: src, match: "string"})
			expect(typs.Spans[1]).To(matchPosition{src: src, match: "string", idx: 1})
		})
	})

	o.Group("Slice", func() {
		const src = `
		package foo
		
		func main() {
			var v []string
			u := v[:]
			x := []string{"foo"}
			y := x[1:]
			z := x[:1]
		}`

		o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
			s := syntax.New()
			err := s.Parse(src)
			expect(err).To(BeNil())

			return expect, s.Layers()
		})

		o.Spec("it highlights brackets and braces for slices within a nested scope", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			arrParens := findLayer(theme.ScopePair+1, layers)
			expect(arrParens.Spans).To(HaveLen(12))

			expect(arrParens.Spans[0]).To(matchPosition{src: src, match: "["})
			expect(arrParens.Spans[1]).To(matchPosition{src: src, match: "]"})
			expect(arrParens.Spans[2]).To(matchPosition{src: src, match: "[", idx: 1})
			expect(arrParens.Spans[3]).To(matchPosition{src: src, match: "]", idx: 1})
			expect(arrParens.Spans[4]).To(matchPosition{src: src, match: "[", idx: 2})
			expect(arrParens.Spans[5]).To(matchPosition{src: src, match: "]", idx: 2})
			expect(arrParens.Spans[6]).To(matchPosition{src: src, match: "{", idx: 1})
			expect(arrParens.Spans[7]).To(matchPosition{src: src, match: "}"})
			expect(arrParens.Spans[8]).To(matchPosition{src: src, match: "[", idx: 3})
			expect(arrParens.Spans[9]).To(matchPosition{src: src, match: "]", idx: 3})
			expect(arrParens.Spans[10]).To(matchPosition{src: src, match: "[", idx: 4})
			expect(arrParens.Spans[11]).To(matchPosition{src: src, match: "]", idx: 4})
		})
	})
}
