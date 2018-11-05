// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"testing"

	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"

	"github.com/apoydence/onpar"
	"github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
)

func TestStmt(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) expect.Expectation {
		return expect.New(t)
	})

	o.Group("Assign", func() {
		const src = `
		package foo

		func main() {
			x := 0.1
			y = "foo"
		}`

		o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
			s := syntax.New()
			err := s.Parse(src)
			expect(err).To(BeNil())

			layers := s.Layers()
			expect(layers).To(HaveLen(6))

			return expect, layers
		})

		o.Spec("it has a number layer for 0.1", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			nums := findLayer(theme.Num, layers)
			expect(nums.Spans).To(HaveLen(1))
			expect(nums.Spans[0]).To(matchPosition{src: src, match: "0.1"})
		})

		o.Spec(`it has a string layer for "foo"`, func(expect expect.Expectation, layers []input.SyntaxLayer) {
			strings := findLayer(theme.String, layers)
			expect(strings.Spans).To(HaveLen(1))
			expect(strings.Spans[0]).To(matchPosition{src: src, match: `"foo"`})
		})
	})

	o.Group("CommClause", func() {
		const src = `
		package foo
		
		func main() {
			switch {
			case foo == "bar":
				x = 1
			case y == false:
			default:
				println("bacon")
			}
		}`

		o.BeforeEach(func(expect expect.Expectation) (expect.Expectation, []input.SyntaxLayer) {
			s := syntax.New()
			err := s.Parse(src)
			expect(err).To(BeNil())

			layers := s.Layers()
			expect(layers).To(HaveLen(9))
			return expect, layers
		})

		o.Spec("it has keyword spans for the cases", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			keywords := findLayer(theme.Keyword, layers)
			expect(keywords.Spans).To(HaveLen(6))
			expect(keywords.Spans[3]).To(matchPosition{src: src, match: "case"})
			expect(keywords.Spans[4]).To(matchPosition{src: src, match: "case", idx: 1})
			expect(keywords.Spans[5]).To(matchPosition{src: src, match: "default"})
		})

		o.Spec("it has string spans for the string cases", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			strings := findLayer(theme.String, layers)
			expect(strings.Spans).To(HaveLen(2))
			expect(strings.Spans[0]).To(matchPosition{src: src, match: `"bar"`})
			expect(strings.Spans[1]).To(matchPosition{src: src, match: `"bacon"`})
		})

		o.Spec("it has num spans for the numeric cases", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			nums := findLayer(theme.Num, layers)
			expect(nums.Spans).To(HaveLen(1))
			expect(nums.Spans[0]).To(matchPosition{src: src, match: "1"})
		})

		o.Spec("it has builtin spans for the builtin functions", func(expect expect.Expectation, layers []input.SyntaxLayer) {
			builtins := findLayer(theme.Builtin, layers)
			expect(builtins.Spans).To(HaveLen(1))
			expect(builtins.Spans[0]).To(matchPosition{src: src, match: "println"})
		})
	})
}
