// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import (
	"fmt"
	"testing"

	"github.com/apoydence/onpar"
	"github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/syntax"
	"github.com/nelsam/vidar/theme"
)

func TestKeywords(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	const src = `
	package foo

	import "time"

	var f map[string]string

	const s = 0

	type foo struct {}
	
	type bar interface{}
	
	func Foo(in chan string) string {
		wait := make(chan struct{})
		go func() {
			var done chan<- struct{} = wait
			defer close(done)
			for range in {
			}
		}()
		var exitAfter <-chan struct{} = wait
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
	}`

	o.BeforeEach(func(t *testing.T) (expect.Expectation, []text.SyntaxLayer) {
		expect := expect.New(t)

		s := syntax.New()
		err := s.Parse(src)
		expect(err).To(BeNil())

		return expect, s.Layers()
	})

	matchers := []matchPosition{
		{src: src, match: "package"},
		{src: src, match: "import"},
		{src: src, match: "var"},
		{src: src, match: "map"},
		{src: src, match: "const"},
		{src: src, match: "type"},
		{src: src, match: "struct"},
		{src: src, idx: 1, match: "type"},
		{src: src, match: "interface"},
		{src: src, match: "func"},
		{src: src, match: "chan"},
		{src: src, idx: 1, match: "chan"},
		{src: src, idx: 1, match: "struct"},
		{src: src, match: "go"},
		{src: src, idx: 1, match: "func"},
		{src: src, idx: 1, match: "var"},
		{src: src, match: "chan<-"},
		{src: src, idx: 2, match: "struct"},
		{src: src, match: "defer"},
		{src: src, match: "for"},
		{src: src, match: "range"},
		{src: src, idx: 2, match: "var"},
		{src: src, match: "<-chan"},
		{src: src, idx: 3, match: "struct"},
		{src: src, match: "goto"},
		{src: src, idx: 1, match: "for"},
		{src: src, match: "if"},
		{src: src, match: "break"},
		{src: src, match: "else"},
		{src: src, idx: 1, match: "if"},
		{src: src, match: "continue"},
		{src: src, idx: 1, match: "else"},
		{src: src, match: "switch"},
		{src: src, match: "case"},
		{src: src, match: "fallthrough"},
		{src: src, match: "default"},
		{src: src, match: "return"},
		{src: src, match: "select"},
		{src: src, idx: 1, match: "case"},
		{src: src, idx: 1, match: "default"},
	}

	for idx, match := range matchers {
		name := fmt.Sprintf("%d%s %s", match.idx+1, numSuffix(match.idx+1), match.match)
		o.Spec(name, func(expect expect.Expectation, layers []text.SyntaxLayer) {
			keywords := findLayer(theme.Keyword, layers)
			expect(keywords.Spans[idx]).To(match)
		})
	}
}
