// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import "go/ast"

func (s *Syntax) addUnresolved(unresolved *ast.Ident) {
	switch unresolved.String() {
	case "append", "cap", "close", "complex", "copy",
		"delete", "imag", "len", "make", "new", "panic",
		"print", "println", "real", "recover":

		s.addNode(s.Theme.Colors.Builtin, unresolved)
	case "nil":
		s.addNode(s.Theme.Colors.Nil, unresolved)
	}
}
