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

func TestParens(t *testing.T) {
	t.Run("FunctionBody", FuncBody)
	t.Run("Array", Array)
	t.Run("Slice", Slice)
}

func FuncBody(t *testing.T) {
	expect := expect.New(t)

	src := `
	package foo

	func (*Foo) Foo(baz Baz) {
		baz.Bar()
	}
	`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil().Else.FailNow()

	layers := s.Layers()

	firstParens := findLayer(theme.ScopePair, layers)
	expect(firstParens.Spans).To.Have.Len(6).Else.FailNow()

	expect(firstParens.Spans[0]).To.Pass(position{src: src, match: "("})
	expect(firstParens.Spans[1]).To.Pass(position{src: src, match: ")"})
	expect(firstParens.Spans[2]).To.Pass(position{src: src, match: "(", idx: 1})
	expect(firstParens.Spans[3]).To.Pass(position{src: src, match: ")", idx: 1})
	expect(firstParens.Spans[4]).To.Pass(position{src: src, match: "{"})
	expect(firstParens.Spans[5]).To.Pass(position{src: src, match: "}"})

	secondParens := findLayer(theme.ScopePair+1, layers)
	expect(secondParens.Spans).To.Have.Len(2).Else.FailNow()
	expect(secondParens.Spans[0]).To.Pass(position{src: src, match: "(", idx: 2})
	expect(secondParens.Spans[1]).To.Pass(position{src: src, match: ")", idx: 2})
}

func Array(t *testing.T) {
	expect := expect.New(t)

	src := `
	package foo
	
	func main() {
		var v [3]string
		v := [...]string{"foo"}
	}
	`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil().Else.FailNow()

	layers := s.Layers()

	// Skip the function ScopePair
	arrParens := findLayer(theme.ScopePair+1, layers)
	expect(arrParens.Spans).To.Have.Len(6).Else.FailNow()
	expect(arrParens.Spans[0]).To.Pass(position{src: src, match: "["})
	expect(arrParens.Spans[1]).To.Pass(position{src: src, match: "]"})
	expect(arrParens.Spans[2]).To.Pass(position{src: src, match: "[", idx: 1})
	expect(arrParens.Spans[3]).To.Pass(position{src: src, match: "]", idx: 1})
	expect(arrParens.Spans[4]).To.Pass(position{src: src, match: "{", idx: 1})
	expect(arrParens.Spans[5]).To.Pass(position{src: src, match: "}"})

	typs := findLayer(theme.Type, layers)
	expect(typs.Spans).To.Have.Len(2).Else.FailNow()
	expect(typs.Spans[0]).To.Pass(position{src: src, match: "string"})
	expect(typs.Spans[1]).To.Pass(position{src: src, match: "string", idx: 1})
}

func Slice(t *testing.T) {
	expect := expect.New(t)

	src := `
	package foo
	
	func main() {
		var v [3]string
		u := v[:]
		x := []string{"foo"}
		y := x[1:]
		z := x[:1]
	}
	`
	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil().Else.FailNow()

	layers := s.Layers()

	// Skip the function ScopePair
	arrParens := findLayer(theme.ScopePair+1, layers)
	expect(arrParens.Spans).To.Have.Len(12).Else.FailNow()
	// 0 and 1 are the [3]string, for an array type.
	expect(arrParens.Spans[2]).To.Pass(position{src: src, match: "[", idx: 1})
	expect(arrParens.Spans[3]).To.Pass(position{src: src, match: "]", idx: 1})
	expect(arrParens.Spans[4]).To.Pass(position{src: src, match: "[", idx: 2})
	expect(arrParens.Spans[5]).To.Pass(position{src: src, match: "]", idx: 2})
	expect(arrParens.Spans[6]).To.Pass(position{src: src, match: "{", idx: 1})
	expect(arrParens.Spans[7]).To.Pass(position{src: src, match: "}"})
	expect(arrParens.Spans[8]).To.Pass(position{src: src, match: "[", idx: 3})
	expect(arrParens.Spans[9]).To.Pass(position{src: src, match: "]", idx: 3})
	expect(arrParens.Spans[10]).To.Pass(position{src: src, match: "[", idx: 4})
	expect(arrParens.Spans[11]).To.Pass(position{src: src, match: "]", idx: 4})
}
