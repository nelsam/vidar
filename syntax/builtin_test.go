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

func TestBuiltins(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	const src = `
	package foo

	func foo() {
		defer recover()

		append([]string{"foo"}, "bar")
		cap([]string{"foo"})
		copy([]string{""}, []string{"foo"})
		delete(map[string]string{"foo": "bar"}, "foo")
		len([]string{""})
		new(struct{})
		print("foo")
		println("foo")

		ch := make(chan struct{})
		close(ch)

		comp := complex(1, -1)
		imag(comp)
		real(comp)

		v := nil

		panic("foo")
	}`

	o.BeforeEach(func(t *testing.T) (expect.Expectation, []input.SyntaxLayer) {
		expect := expect.New(t)

		s := syntax.New()
		err := s.Parse(src)
		expect(err).To(BeNil())

		layers := s.Layers()

		return expect, layers
	})

	o.Spec("it highlights the expected builtins", func(expect expect.Expectation, layers []input.SyntaxLayer) {
		builtins := findLayer(theme.Builtin, layers)
		expect(builtins.Spans[0]).To(matchPosition{src: src, match: "recover"})
		expect(builtins.Spans[1]).To(matchPosition{src: src, match: "append"})
		expect(builtins.Spans[2]).To(matchPosition{src: src, match: "cap"})
		expect(builtins.Spans[3]).To(matchPosition{src: src, match: "copy"})
		expect(builtins.Spans[4]).To(matchPosition{src: src, match: "delete"})
		expect(builtins.Spans[5]).To(matchPosition{src: src, match: "len"})
		expect(builtins.Spans[6]).To(matchPosition{src: src, match: "new"})
		expect(builtins.Spans[7]).To(matchPosition{src: src, match: "print"})
		expect(builtins.Spans[8]).To(matchPosition{src: src, match: "println"})
		expect(builtins.Spans[9]).To(matchPosition{src: src, match: "make"})
		expect(builtins.Spans[10]).To(matchPosition{src: src, match: "close"})
		expect(builtins.Spans[11]).To(matchPosition{src: src, match: "complex"})
		expect(builtins.Spans[12]).To(matchPosition{src: src, match: "imag"})
		expect(builtins.Spans[13]).To(matchPosition{src: src, match: "real"})
		expect(builtins.Spans[14]).To(matchPosition{src: src, match: "panic"})
	})

	o.Spec("it highlights nil", func(expect expect.Expectation, layers []input.SyntaxLayer) {
		nils := findLayer(theme.Nil, layers)
		expect(nils.Spans).To(HaveLen(1))
		expect(nils.Spans[0]).To(matchPosition{src: src, match: "nil"})
	})

	o.Spec("it does not highlight builtins as idents", func(expect expect.Expectation, layers []input.SyntaxLayer) {
		idents := findLayer(theme.Ident, layers)
		expect(idents.Spans).To(HaveLen(6))
	})
}
