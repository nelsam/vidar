// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"
	"github.com/poy/onpar"
	"github.com/poy/onpar/expect"
	"github.com/poy/onpar/matchers"
)

var (
	equal = matchers.Equal
	not   = matchers.Not
)

func TestGeneric(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (expect.Expectation, syntax.Generic) {
		return expect.New(t), syntax.Generic{
			Scopes: []syntax.Scope{{Open: "{", Close: "}"}},
			Wrapped: []syntax.Wrapped{
				{Open: "//", Close: "\n", Construct: theme.Comment},
				{Open: "'", Close: "'", Escapes: []string{`\'`}, Construct: theme.String},
				{Open: "/*", Close: "*/", Nested: true, Construct: theme.Comment},
			},
			StaticWords: map[string]theme.LanguageConstruct{
				"package": theme.Keyword,
				"make":    theme.Builtin,
			},
			DynamicWords: []syntax.Word{
				{Before: "fn", Construct: theme.Func},
				{Before: "var", After: "int", Construct: theme.Num},
			},
		}
	})

	o.Spec("it understands scope", func(expect expect.Expectation, g syntax.Generic) {
		m := g.Parse([]rune(` {  {  }}`))
		expect(m.Depth(0)).To(equal(0))
		expect(m.Depth(1)).To(equal(1))
		expect(m.Depth(4)).To(equal(2))
		expect(m.Depth(8)).To(equal(1))
	})

	o.Group("syntax layers", func() {
		o.BeforeEach(func(expect expect.Expectation, g syntax.Generic) (expect.Expectation, []input.SyntaxLayer, string) {
			source := `
				// some docs
				foo := 'some string\' with escapes'
				/* and comment blocks /* can nest */ without closing early */
				{ will open a new scope
					package is a keyword
					numbers like 12.3 and -5 should be highlighted as numeric, always
					numbers like 1.2.3 should _not_ be highlighted as numeric, though
					{ will open a nested scope and should be rainbow highlighted
						make is a builtin
					} should match the theme.LanguageConstruct value of the brace it closes
				}
				fn somefunc is how we define a function
				we highlight variables by their type with var somevariable int`
			m := g.Parse([]rune(source))
			return expect, m.Layers(), source
		})

		o.Spec("it recognizes single-line comments", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(haveLayer(theme.Comment, source, "// some docs\n"))
		})

		o.Spec("it recognizes nested comment blocks", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(haveLayer(theme.Comment, source, "/* and comment blocks /* can nest */ without closing early */"))
		})

		o.Spec("it recognizes strings and skips escaped quotes", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(haveLayer(theme.String, source, `'some string\' with escapes'`))
		})

		o.Spec("it recognizes positive and negative numbers", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(haveLayer(theme.Num, source, "12.3"))
			expect(layers).To(haveLayer(theme.Num, source, "-5"))
		})

		o.Spec("it does not highlight semantic versions", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(not(haveLayer(theme.Num, source, "1.2.3")))
		})

		o.Spec("it recognizes static words", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(haveLayer(theme.Keyword, source, "package"))
			expect(layers).To(haveLayer(theme.Builtin, source, "make"))
		})

		o.Spec("it recognizes dynamic words surrounded by static words", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(haveLayer(theme.Func, source, "somefunc"))
			expect(layers).To(haveLayer(theme.Num, source, "somevariable"))
		})

		o.Spec("it handles rainbow colors for scope pairs", func(expect expect.Expectation, layers []input.SyntaxLayer, source string) {
			expect(layers).To(haveLayer(theme.ScopePair, source, "{"))
			expect(layers).To(haveLayer(theme.ScopePair+1, source, "{", nth(2)))
			expect(layers).To(haveLayer(theme.ScopePair+1, source, "}"))
			expect(layers).To(haveLayer(theme.ScopePair, source, "}", nth(2)))
		})
	})
}

type haveLayerMatcher struct {
	construct        theme.LanguageConstruct
	source, expected string
	skip             int
}

type layerOpt func(haveLayerMatcher) haveLayerMatcher

// nth returns a layerOpt that tells the haveLayerMatcher to match the nth
// occurrance of the expected string.
func nth(n int) layerOpt {
	return func(m haveLayerMatcher) haveLayerMatcher {
		m.skip = n - 1
		return m
	}
}

func haveLayer(construct theme.LanguageConstruct, source, expected string, opts ...layerOpt) haveLayerMatcher {
	m := haveLayerMatcher{construct: construct, source: source, expected: expected}
	for _, o := range opts {
		m = o(m)
	}
	return m
}

func (m haveLayerMatcher) Match(actual interface{}) (interface{}, error) {
	start := indexN(m.source, m.expected, m.skip)
	end := start + len(m.expected)
	switch l := actual.(type) {
	case input.SyntaxLayer:
		return actual, m.matchLayer(l, start, end)
	case []input.SyntaxLayer:
		for _, layer := range l {
			if err := m.matchLayer(layer, start, end); err == nil {
				return actual, nil
			}
		}
		return actual, fmt.Errorf("expected to find a span from %d to %d in a layer with construct %#v in %#v", start, end, m.construct, l)
	default:
		return actual, fmt.Errorf("expected actual to be of type input.SyntaxLayer or []input.SyntaxLayer; was %T", actual)
	}
}

func (m haveLayerMatcher) matchLayer(l input.SyntaxLayer, start, end int) error {
	if l.Construct != m.construct {
		return fmt.Errorf("expected construct %#v to equal %#v", l.Construct, m.construct)
	}
	for _, s := range l.Spans {
		if s.Start == start && s.End == end {
			return nil
		}
	}
	return fmt.Errorf("expected to find a span from %d to %d in %#v", start, end, l.Spans)
}

func indexN(haystack, needle string, n int) int {
	i := strings.Index(haystack, needle)
	if i == -1 {
		return -1
	}
	for ; n > 0; n-- {
		i += len(needle)
		nextIdx := strings.Index(haystack[i:], needle)
		if nextIdx == -1 {
			return -1
		}
		i += nextIdx
	}
	return i
}
