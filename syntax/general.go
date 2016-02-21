// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"go/ast"
	"go/token"

	"github.com/nelsam/gxui"
)

func handleFieldList(src *ast.FieldList) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(src.List))
	if src.Opening != 0 {
		layers = append(layers, layer(src.Opening, 1, defaultRainbow.New()))
	}
	for _, block := range src.List {
		layers = append(layers, nodeLayer(block.Type, typeColor))
	}
	if src.Closing != 0 {
		layers = append(layers, layer(src.Closing, 1, defaultRainbow.Pop()))
	}
	return layers
}

func handleStructType(src *ast.StructType) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 1+len(src.Fields.List))
	layers = append(layers, layer(src.Struct, len("struct"), keywordColor))
	layers = append(layers, handleFieldList(src.Fields)...)
	return layers
}

func handleFuncType(src *ast.FuncType) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	if src.Func != token.NoPos {
		layers = append(layers, layer(src.Func, len("func"), keywordColor))
	}
	if src.Params != nil {
		layers = append(layers, handleFieldList(src.Params)...)
	}
	if src.Results != nil {
		layers = append(layers, handleFieldList(src.Results)...)
	}
	return layers
}

func handleInterfaceType(src *ast.InterfaceType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{layer(src.Interface, len("interface"), keywordColor)}
	layers = append(layers, handleFieldList(src.Methods)...)
	return layers
}

func handleMapType(src *ast.MapType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{layer(src.Map, len("map"), keywordColor)}
	layers = append(layers, handleExpr(src.Key)...)
	layers = append(layers, handleExpr(src.Value)...)
	return layers
}

func handleArrayType(src *ast.ArrayType) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{layer(src.Lbrack, 1, defaultRainbow.New())}
	layers = append(layers, handleExpr(src.Len)...)
	layers = append(layers, layer(src.End(), 1, defaultRainbow.Pop()))
	layers = append(layers, handleExpr(src.Elt)...)
	return layers
}

func handleBasicLit(src *ast.BasicLit) gxui.CodeSyntaxLayers {
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

func handleCompositeLit(src *ast.CompositeLit) gxui.CodeSyntaxLayers {
	layers := append(
		handleExpr(src.Type),
		layer(src.Lbrace, 1, defaultRainbow.New()),
	)
	for _, elt := range src.Elts {
		layers = append(layers, handleExpr(elt)...)
	}
	return append(layers, layer(src.Rbrace, 1, defaultRainbow.Pop()))
}
