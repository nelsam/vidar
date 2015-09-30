// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"fmt"
	"go/ast"

	"github.com/google/gxui"
)

func handleExpr(expr ast.Expr) gxui.CodeSyntaxLayers {
	switch expr.(type) {
	case *ast.BadExpr:
		return nil
	case *ast.BinaryExpr:
		return nil
	case *ast.CallExpr:
		return nil
	case *ast.IndexExpr:
		return nil
	case *ast.KeyValueExpr:
		return nil
	case *ast.ParenExpr:
		return nil
	case *ast.SelectorExpr:
		return nil
	case *ast.SliceExpr:
		return nil
	case *ast.StarExpr:
		return nil
	case *ast.TypeAssertExpr:
		return nil
	case *ast.UnaryExpr:
		return nil
	case *ast.Ident:
		return nil
	case *ast.CompositeLit:
		return nil
	default:
		panic(fmt.Errorf("Unknown expression type: %T", expr))
	}
}
