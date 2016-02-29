// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"log"

	"github.com/nelsam/gxui"
)

func (l layers) handleExpr(expr ast.Expr) gxui.CodeSyntaxLayers {
	if expr == nil {
		return nil
	}
	switch src := expr.(type) {
	case *ast.ArrayType:
		return l.handleArrayType(src)
	case *ast.BadExpr:
		return l.handleBadExpr(src)
	case *ast.BasicLit:
		return l.handleBasicLit(src)
	case *ast.BinaryExpr:
		return l.handleBinaryExpr(src)
	case *ast.CallExpr:
		return l.handleCallExpr(src)
	case *ast.ChanType:
		return l.handleChanExpr(src)
	case *ast.FuncLit:
		return l.handleFuncLitExpr(src)
	case *ast.FuncType:
		return l.handleFuncType(src)
	case *ast.IndexExpr:
		return l.handleIndexExpr(src)
	case *ast.InterfaceType:
		return l.handleInterfaceType(src)
	case *ast.KeyValueExpr:
		return l.handleKeyValueExpr(src)
	case *ast.MapType:
		return l.handleMapType(src)
	case *ast.ParenExpr:
		return l.handleParenExpr(src)
	case *ast.SelectorExpr:
		return l.handleSelectorExpr(src)
	case *ast.SliceExpr:
		return l.handleSliceExpr(src)
	case *ast.StructType:
		return l.handleStructType(src)
	case *ast.StarExpr:
		return l.handleStarExpr(src)
	case *ast.TypeAssertExpr:
		return l.handleTypeAssertExpr(src)
	case *ast.UnaryExpr:
		return l.handleUnaryExpr(src)
	case *ast.Ident:
		return nil
	case *ast.CompositeLit:
		return l.handleCompositeLit(src)
	default:
		log.Printf("Error: Unknown expression type: %T", expr)
		return nil
	}
}

func (l layers) handleBadExpr(src *ast.BadExpr) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{l.nodeLayer(src, badColor, badBackground)}
}

func (l layers) handleBinaryExpr(src *ast.BinaryExpr) gxui.CodeSyntaxLayers {
	return append(l.handleExpr(src.X), l.handleExpr(src.Y)...)
}

func (l layers) handleCallExpr(src *ast.CallExpr) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(src.Args)+1)
	layers = append(layers, l.handleExpr(src.Fun)...)
	layers = append(layers, l.layer(src.Lparen, 1, defaultRainbow.New()))
	for _, arg := range src.Args {
		layers = append(layers, l.handleExpr(arg)...)
	}
	return append(layers, l.layer(src.Rparen, 1, defaultRainbow.Pop()))
}

func (l layers) handleChanExpr(src *ast.ChanType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{l.layer(src.Begin, len("chan"), keywordColor)}
	layers = append(layers, l.handleExpr(src.Value)...)
	return layers
}

func (l layers) handleFuncLitExpr(src *ast.FuncLit) gxui.CodeSyntaxLayers {
	return append(l.handleFuncType(src.Type), l.handleBlockStmt(src.Body)...)
}

func (l layers) handleIndexExpr(src *ast.IndexExpr) gxui.CodeSyntaxLayers {
	layers := l.handleExpr(src.X)
	layers = append(layers, l.layer(src.Lbrack, 1, defaultRainbow.New()))
	layers = append(layers, l.handleExpr(src.Index)...)
	return append(layers, l.layer(src.Rbrack, 1, defaultRainbow.Pop()))
}

func (l layers) handleKeyValueExpr(src *ast.KeyValueExpr) gxui.CodeSyntaxLayers {
	return append(l.handleExpr(src.Key), l.handleExpr(src.Value)...)
}

func (l layers) handleParenExpr(src *ast.ParenExpr) gxui.CodeSyntaxLayers {
	parens := gxui.CodeSyntaxLayers{l.layer(src.Lparen, 1, defaultRainbow.New())}
	parens = append(parens, l.handleExpr(src.X)...)
	return append(parens, l.layer(src.Rparen, 1, defaultRainbow.Pop()))
}

func (l layers) handleSelectorExpr(src *ast.SelectorExpr) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{l.nodeLayer(src.Sel, functionColor)}
}

func (l layers) handleSliceExpr(src *ast.SliceExpr) gxui.CodeSyntaxLayers {
	layers := l.handleExpr(src.X)
	layers = append(layers, l.layer(src.Lbrack, 1, defaultRainbow.New()))
	layers = append(layers, l.handleExpr(src.Low)...)
	layers = append(layers, l.handleExpr(src.High)...)
	layers = append(layers, l.handleExpr(src.Max)...)
	return append(layers, l.layer(src.Lbrack, 1, defaultRainbow.Pop()))
}

func (l layers) handleStarExpr(src *ast.StarExpr) gxui.CodeSyntaxLayers {
	return l.handleExpr(src.X)
}

func (l layers) handleTypeAssertExpr(src *ast.TypeAssertExpr) gxui.CodeSyntaxLayers {
	layers := l.handleExpr(src.X)
	layers = append(
		layers,
		l.layer(src.Lparen, 1, defaultRainbow.New()),
	)
	layers = append(layers, l.handleExpr(src.Type)...)
	return append(layers, l.layer(src.Rparen, 1, defaultRainbow.Pop()))
}

func (l layers) handleUnaryExpr(src *ast.UnaryExpr) gxui.CodeSyntaxLayers {
	return l.handleExpr(src.X)
}
