// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"testing"

	"github.com/apoydence/onpar"
	"github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/plugin/gosyntax/syntax"
	"github.com/nelsam/vidar/theme"
)

func TestUnicode(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	const src = `
	package foo
	
	func µ() string {
		var þ = "Ωð"
		return þ
	}`

	o.BeforeEach(func(t *testing.T) (expect.Expectation, []input.SyntaxLayer) {
		expect := expect.New(t)

		s := syntax.New()
		err := s.Parse(src)
		expect(err).To(BeNil())
		layers := s.Layers()

		return expect, layers
	})

	o.Spec("it sets keyword spans to the correct index with unicode characters", func(expect expect.Expectation, layers []input.SyntaxLayer) {
		keywords := findLayer(theme.Keyword, layers)
		expect(keywords.Spans).To(HaveLen(4))
		expect(keywords.Spans[2]).To(matchPosition{src: src, match: "var"})
		expect(keywords.Spans[3]).To(matchPosition{src: src, match: "return"})
	})

	o.Spec("it sets string spans to the correct index with unicode characters", func(expect expect.Expectation, layers []input.SyntaxLayer) {
		strings := findLayer(theme.String, layers)
		expect(strings.Spans).To(HaveLen(1))
		expect(strings.Spans[0]).To(matchPosition{src: src, match: `"Ωð"`})
	})
}

func TestPackageDocs(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	// Note: leading whitespace does affect the test, so leave this unindented.
	const src = `
// Package foo does stuff.
// It is also a thing.
package foo`

	o.BeforeEach(func(t *testing.T) (expect.Expectation, []input.SyntaxLayer) {
		expect := expect.New(t)

		s := syntax.New()
		err := s.Parse(src)
		expect(err).To(BeNil())

		layers := s.Layers()
		expect(layers).To(HaveLen(2))

		return expect, layers
	})

	o.Spec("it highlights package docs", func(expect expect.Expectation, layers []input.SyntaxLayer) {
		comments := findLayer(theme.Comment, layers)
		expect(comments.Spans).To(HaveLen(1))
		comment := "// Package foo does stuff.\n" +
			"// It is also a thing."
		expect(comments.Spans[0]).To(matchPosition{src: src, match: comment})
	})
}
