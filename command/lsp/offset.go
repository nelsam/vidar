// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package lsp

import (
	"github.com/nelsam/vidar/commander/input"
	"github.com/sourcegraph/go-lsp"
)

func Offset(e input.Editor, p lsp.Position) int {
	return e.LineStart(p.Line) + p.Character
}
