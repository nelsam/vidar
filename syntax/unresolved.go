// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import "go/ast"

func (s *Syntax) addUnresolved(unresolved *ast.Ident) {
	// This used to be used to add builtins, but that's
	// now handled by the *ast.Ident case in addExpr.
}
