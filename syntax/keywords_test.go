// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/nelsam/vidar/syntax"
)

func TestKeywords(t *testing.T) {
	expect := expect.New(t)

	src := `
	package foo

	import "time"

	var f map[string]string

	const s = 0

	type foo struct {}
	
	type bar interface{}
	
	func Foo(in chan string) string {
		done := make(chan struct{})
		go func() {
			defer close(done)
			for range in {
			}
		}()
		goto FOO

FOO:
		for {
			if true {
				break
			} else if false {
				continue
			} else {
				panic("foo")
			}
		}
		switch {
		case true:
			fallthrough
		default:
			return "foo"
		}
		<-done
		select {
		case <-time.After(time.Second):
		default:
		}
	}
	`
	s := syntax.New(syntax.DefaultTheme)
	err := s.Parse(src)
	expect(err).To.Be.Nil().Else.FailNow()

	layers := s.Layers()

	keywords := layers[syntax.DefaultTheme.Colors.Keyword]
	expect(keywords.Spans()).To.Have.Len(34).Else.FailNow()

	expect(keywords.Spans()[0]).To.Pass(position{src: src, match: "package"})
	expect(keywords.Spans()[1]).To.Pass(position{src: src, match: "import"})
	expect(keywords.Spans()[2]).To.Pass(position{src: src, match: "var"})
	expect(keywords.Spans()[3]).To.Pass(position{src: src, match: "map"})
	expect(keywords.Spans()[4]).To.Pass(position{src: src, match: "const"})
	expect(keywords.Spans()[5]).To.Pass(position{src: src, match: "type"})
	expect(keywords.Spans()[6]).To.Pass(position{src: src, match: "struct"})
	expect(keywords.Spans()[7]).To.Pass(position{src: src, idx: 1, match: "type"})
	expect(keywords.Spans()[8]).To.Pass(position{src: src, match: "interface"})
	expect(keywords.Spans()[9]).To.Pass(position{src: src, match: "func"})
	expect(keywords.Spans()[10]).To.Pass(position{src: src, match: "chan"})
	expect(keywords.Spans()[11]).To.Pass(position{src: src, idx: 1, match: "chan"})
	expect(keywords.Spans()[12]).To.Pass(position{src: src, idx: 1, match: "struct"})
	expect(keywords.Spans()[13]).To.Pass(position{src: src, match: "go"})
	expect(keywords.Spans()[14]).To.Pass(position{src: src, idx: 1, match: "func"})
	expect(keywords.Spans()[15]).To.Pass(position{src: src, match: "defer"})
	expect(keywords.Spans()[16]).To.Pass(position{src: src, match: "for"})
	expect(keywords.Spans()[17]).To.Pass(position{src: src, match: "range"})
	expect(keywords.Spans()[18]).To.Pass(position{src: src, match: "goto"})
	expect(keywords.Spans()[19]).To.Pass(position{src: src, idx: 1, match: "for"})
	expect(keywords.Spans()[20]).To.Pass(position{src: src, match: "if"})
	expect(keywords.Spans()[21]).To.Pass(position{src: src, match: "break"})
	expect(keywords.Spans()[22]).To.Pass(position{src: src, match: "else"})
	expect(keywords.Spans()[23]).To.Pass(position{src: src, idx: 1, match: "if"})
	expect(keywords.Spans()[24]).To.Pass(position{src: src, match: "continue"})
	expect(keywords.Spans()[25]).To.Pass(position{src: src, idx: 1, match: "else"})
	expect(keywords.Spans()[26]).To.Pass(position{src: src, match: "switch"})
	expect(keywords.Spans()[27]).To.Pass(position{src: src, match: "case"})
	expect(keywords.Spans()[28]).To.Pass(position{src: src, match: "fallthrough"})
	expect(keywords.Spans()[29]).To.Pass(position{src: src, match: "default"})
	expect(keywords.Spans()[30]).To.Pass(position{src: src, match: "return"})
	expect(keywords.Spans()[31]).To.Pass(position{src: src, match: "select"})
	expect(keywords.Spans()[32]).To.Pass(position{src: src, idx: 1, match: "case"})
	expect(keywords.Spans()[33]).To.Pass(position{src: src, idx: 1, match: "default"})
}
