// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"go/token"

	"github.com/nelsam/gxui"
)

func (l layers) handleFieldList(src *ast.FieldList) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(src.List))
	if src.Opening != 0 {
		layers = append(layers, l.layer(src.Opening, 1, defaultRainbow.New()))
	}
	for _, block := range src.List {
		layers = append(layers, l.nodeLayer(block.Type, typeColor))
	}
	if src.Closing != 0 {
		layers = append(layers, l.layer(src.Closing, 1, defaultRainbow.Pop()))
	}
	return layers
}

func (l layers) handleStructType(src *ast.StructType) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 1+len(src.Fields.List))
	layers = append(layers, l.layer(src.Struct, len("struct"), keywordColor))
	layers = append(layers, l.handleFieldList(src.Fields)...)
	return layers
}

func (l layers) handleFuncType(src *ast.FuncType) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	if src.Func != token.NoPos {
		layers = append(layers, l.layer(src.Func, len("func"), keywordColor))
	}
	if src.Params != nil {
		layers = append(layers, l.handleFieldList(src.Params)...)
	}
	if src.Results != nil {
		layers = append(layers, l.handleFieldList(src.Results)...)
	}
	return layers
}

func (l layers) handleInterfaceType(src *ast.InterfaceType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{l.layer(src.Interface, len("interface"), keywordColor)}
	layers = append(layers, l.handleFieldList(src.Methods)...)
	return layers
}

func (l layers) handleMapType(src *ast.MapType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{l.layer(src.Map, len("map"), keywordColor)}
	layers = append(layers, l.handleExpr(src.Key)...)
	layers = append(layers, l.handleExpr(src.Value)...)
	return layers
}

func (l layers) handleArrayType(src *ast.ArrayType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{
		l.layer(src.Lbrack, 1, defaultRainbow.New()),
		l.layer(src.Lbrack+1, 1, defaultRainbow.Pop()),
	}
	layers = append(layers, l.handleExpr(src.Len)...)
	layers = append(layers, l.handleExpr(src.Elt)...)
	return layers
}

func (l layers) handleBasicLit(src *ast.BasicLit) gxui.CodeSyntaxLayers {
	var color gxui.Color
	switch src.Kind {
	case token.INT, token.FLOAT:
		color = numColor
	case token.CHAR, token.STRING:
		color = stringColor
	default:
		return nil
	}
	return gxui.CodeSyntaxLayers{l.nodeLayer(src, color)}
}

func (l layers) handleCompositeLit(src *ast.CompositeLit) gxui.CodeSyntaxLayers {
	layers := append(
		l.handleExpr(src.Type),
		l.layer(src.Lbrace, 1, defaultRainbow.New()),
	)
	for _, elt := range src.Elts {
		layers = append(layers, l.handleExpr(elt)...)
	}
	return append(layers, l.layer(src.Rbrace, 1, defaultRainbow.Pop()))
}

func (l layers) handleCommClause(src *ast.CommClause) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	length := len("case")
	if src.Comm == nil {
		length = len("default")
	}
	layers = append(layers, l.layer(src.Case, length, keywordColor))
	layers = append(layers, l.handleStmt(src.Comm)...)
	for _, stmt := range src.Body {
		layers = append(layers, l.handleStmt(stmt)...)
	}
	return layers
}
