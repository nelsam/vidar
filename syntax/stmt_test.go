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

func TestStmt(t *testing.T) {
	t.Run("Assign", AssignStmt)
	t.Run("CommClause", CommClause)
}

func AssignStmt(t *testing.T) {
	expect := expect.New(t)

	src := `
package foo
	
func main() {
	x := 0.1
	y = "foo"
}`

	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(6)

	nums := findLayer(theme.Num, layers)
	expect(nums.Spans).To.Have.Len(1)
	expect(nums.Spans[0]).To.Pass(position{src: src, match: "0.1"})

	strings := findLayer(theme.String, layers)
	expect(strings.Spans).To.Have.Len(1)
	expect(strings.Spans[0]).To.Pass(position{src: src, match: `"foo"`})
}

func CommClause(t *testing.T) {
	expect := expect.New(t)

	src := `
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

	s := syntax.New()
	err := s.Parse(src)
	expect(err).To.Be.Nil()

	layers := s.Layers()
	expect(layers).To.Have.Len(9)

	keywords := findLayer(theme.Keyword, layers)
	expect(keywords.Spans).To.Have.Len(6)
	expect(keywords.Spans[3]).To.Pass(position{src: src, match: "case"})
	expect(keywords.Spans[4]).To.Pass(position{src: src, match: "case", idx: 1})
	expect(keywords.Spans[5]).To.Pass(position{src: src, match: "default"})

	strings := findLayer(theme.String, layers)
	expect(strings.Spans).To.Have.Len(2)
	expect(strings.Spans[0]).To.Pass(position{src: src, match: `"bar"`})
	expect(strings.Spans[1]).To.Pass(position{src: src, match: `"bacon"`})

	nums := findLayer(theme.Num, layers)
	expect(nums.Spans).To.Have.Len(1)
	expect(nums.Spans[0]).To.Pass(position{src: src, match: "1"})

	builtins := findLayer(theme.Builtin, layers)
	expect(builtins.Spans).To.Have.Len(1)
	expect(builtins.Spans[0]).To.Pass(position{src: src, match: "println"})
}
