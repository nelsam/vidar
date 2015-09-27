// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"go/ast"
	"log"

	"github.com/google/gxui"
)

func handleUnresolved(unresolved *ast.Ident) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 1)
	switch unresolved.String() {
	case "append", "cap", "close", "complex", "copy",
		"delete", "imag", "len", "make", "new", "panic",
		"print", "println", "real", "recover":

		layers = append(layers, nodeLayer(unresolved, builtinColor))
	case "nil":
		layers = append(layers, nodeLayer(unresolved, nilColor))
	default:
		log.Printf("Found unresolved declaration %s at position %d", unresolved.String(), unresolved.Pos())
	}
	return layers
}
