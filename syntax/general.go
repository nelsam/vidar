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
	for _, block := range src.List {
		layers = append(layers, nodeLayer(block.Type, typeColor))
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
	layers := make(gxui.CodeSyntaxLayers, 0, 3)
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
	layers := handleExpr(src.Type)
	for _, elt := range src.Elts {
		layers = append(layers, handleExpr(elt)...)
	}
	return layers
}
