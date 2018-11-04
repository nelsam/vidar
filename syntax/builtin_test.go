// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"
)

func TestBuiltins(t *testing.T) {
	expect := expect.New(t)

	src := `
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
	}
	`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil().Else.FailNow()

	layers := s.Layers()

	builtins := findLayer(theme.Builtin, layers)
	expect(builtins.Spans).To.Have.Len(15).Else.FailNow()

	expect(builtins.Spans[0]).To.Pass(position{src: src, match: "recover"})
	expect(builtins.Spans[1]).To.Pass(position{src: src, match: "append"})
	expect(builtins.Spans[2]).To.Pass(position{src: src, match: "cap"})
	expect(builtins.Spans[3]).To.Pass(position{src: src, match: "copy"})
	expect(builtins.Spans[4]).To.Pass(position{src: src, match: "delete"})
	expect(builtins.Spans[5]).To.Pass(position{src: src, match: "len"})
	expect(builtins.Spans[6]).To.Pass(position{src: src, match: "new"})
	expect(builtins.Spans[7]).To.Pass(position{src: src, match: "print"})
	expect(builtins.Spans[8]).To.Pass(position{src: src, match: "println"})
	expect(builtins.Spans[9]).To.Pass(position{src: src, match: "make"})
	expect(builtins.Spans[10]).To.Pass(position{src: src, match: "close"})
	expect(builtins.Spans[11]).To.Pass(position{src: src, match: "complex"})
	expect(builtins.Spans[12]).To.Pass(position{src: src, match: "imag"})
	expect(builtins.Spans[13]).To.Pass(position{src: src, match: "real"})
	expect(builtins.Spans[14]).To.Pass(position{src: src, match: "panic"})

	nils := findLayer(theme.Nil, layers)
	expect(nils.Spans).To.Have.Len(1).Else.FailNow()
	expect(nils.Spans[0]).To.Pass(position{src: src, match: "nil"})

	// Test that we're not highlighting these as both idents and builtins.
	idents := findLayer(theme.Ident, layers)
	expect(idents.Spans).To.Have.Len(6)
}
