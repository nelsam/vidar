// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/nelsam/gxui"
)

func handleExpr(expr ast.Expr) gxui.CodeSyntaxLayers {
	if expr == nil {
		return nil
	}
	switch src := expr.(type) {
	case *ast.ArrayType:
		return nil
	case *ast.BadExpr:
		return handleBadExpr(src)
	case *ast.BasicLit:
		return handleBasicLitExpr(src)
	case *ast.BinaryExpr:
		return nil
	case *ast.CallExpr:
		return handleCallExpr(src)
	case *ast.ChanType:
		return nil
	case *ast.FuncLit:
		return nil
	case *ast.FuncType:
		return nil
	case *ast.IndexExpr:
		return nil
	case *ast.InterfaceType:
		return nil
	case *ast.KeyValueExpr:
		return nil
	case *ast.MapType:
		return nil
	case *ast.ParenExpr:
		return nil
	case *ast.SelectorExpr:
		return handleSelectorExpr(src)
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

func handleBadExpr(src *ast.BadExpr) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{nodeLayer(src, badColor, badBackground)}
}

func handleBasicLitExpr(src *ast.BasicLit) gxui.CodeSyntaxLayers {
	var color gxui.Color
	switch src.Kind {
	case token.INT, token.FLOAT:
		color = numColor
	case token.CHAR, token.STRING:
		color = stringColor
	default:
		return nil
	}
	return gxui.CodeSyntaxLayers{nodeLayer(src, color)}
}

func handleCallExpr(src *ast.CallExpr) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(src.Args)+1)
	layers = append(layers, handleExpr(src.Fun)...)
	for _, arg := range src.Args {
		layers = append(layers, handleExpr(arg)...)
	}
	return layers
}

func handleSelectorExpr(src *ast.SelectorExpr) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{nodeLayer(src.Sel, functionColor)}
}
