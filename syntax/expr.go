// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"fmt"
	"go/ast"

	"github.com/nelsam/gxui"
)

func handleExpr(expr ast.Expr) gxui.CodeSyntaxLayers {
	if expr == nil {
		return nil
	}
	switch src := expr.(type) {
	case *ast.ArrayType:
		return handleArrayExpr(src)
	case *ast.BadExpr:
		return handleBadExpr(src)
	case *ast.BasicLit:
		return handleBasicLit(src)
	case *ast.BinaryExpr:
		return handleBinaryExpr(src)
	case *ast.CallExpr:
		return handleCallExpr(src)
	case *ast.ChanType:
		return handleChanExpr(src)
	case *ast.FuncLit:
		return handleFuncLitExpr(src)
	case *ast.FuncType:
		return handleFuncType(src)
	case *ast.IndexExpr:
		return handleIndexExpr(src)
	case *ast.InterfaceType:
		return handleInterfaceType(src)
	case *ast.KeyValueExpr:
		return handleKeyValueExpr(src)
	case *ast.MapType:
		return handleMapType(src)
	case *ast.ParenExpr:
		return handleParenExpr(src)
	case *ast.SelectorExpr:
		return handleSelectorExpr(src)
	case *ast.SliceExpr:
		return handleSliceExpr(src)
	case *ast.StarExpr:
		return handleStarExpr(src)
	case *ast.TypeAssertExpr:
		return handleTypeAssertExpr(src)
	case *ast.UnaryExpr:
		return handleUnaryExpr(src)
	case *ast.Ident:
		return nil
	case *ast.CompositeLit:
		return handleCompositeLit(src)
	default:
		panic(fmt.Errorf("Unknown expression type: %T", expr))
	}
}

func handleArrayExpr(src *ast.ArrayType) gxui.CodeSyntaxLayers {
	layers := handleExpr(src.Elt)
	layers = append(layers, handleExpr(src.Len)...)
	return layers
}

func handleBadExpr(src *ast.BadExpr) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{nodeLayer(src, badColor, badBackground)}
}

func handleBinaryExpr(src *ast.BinaryExpr) gxui.CodeSyntaxLayers {
	return append(handleExpr(src.X), handleExpr(src.Y)...)
}

func handleCallExpr(src *ast.CallExpr) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(src.Args)+1)
	layers = append(layers, handleExpr(src.Fun)...)
	for _, arg := range src.Args {
		layers = append(layers, handleExpr(arg)...)
	}
	return layers
}

func handleChanExpr(src *ast.ChanType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{layer(src.Begin, len("chan"), keywordColor)}
	layers = append(layers, handleExpr(src.Value)...)
	return layers
}

func handleFuncLitExpr(src *ast.FuncLit) gxui.CodeSyntaxLayers {
	return append(handleFuncType(src.Type), handleBlockStmt(src.Body)...)
}

func handleIndexExpr(src *ast.IndexExpr) gxui.CodeSyntaxLayers {
	return append(handleExpr(src.X), handleExpr(src.Index)...)
}

func handleKeyValueExpr(src *ast.KeyValueExpr) gxui.CodeSyntaxLayers {
	return append(handleExpr(src.Key), handleExpr(src.Value)...)
}

func handleParenExpr(src *ast.ParenExpr) gxui.CodeSyntaxLayers {
	return handleExpr(src.X)
}

func handleSelectorExpr(src *ast.SelectorExpr) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{nodeLayer(src.Sel, functionColor)}
}

func handleSliceExpr(src *ast.SliceExpr) gxui.CodeSyntaxLayers {
	layers := handleExpr(src.X)
	layers = append(layers, handleExpr(src.Low)...)
	layers = append(layers, handleExpr(src.High)...)
	layers = append(layers, handleExpr(src.Max)...)
	return layers
}

func handleStarExpr(src *ast.StarExpr) gxui.CodeSyntaxLayers {
	return handleExpr(src.X)
}

func handleTypeAssertExpr(src *ast.TypeAssertExpr) gxui.CodeSyntaxLayers {
	return append(handleExpr(src.X), handleExpr(src.Type)...)
}

func handleUnaryExpr(src *ast.UnaryExpr) gxui.CodeSyntaxLayers {
	return handleExpr(src.X)
}
