// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"fmt"
	"go/ast"

	"github.com/nelsam/gxui"
)

func handleDecl(decl ast.Decl) gxui.CodeSyntaxLayers {
	switch src := decl.(type) {
	case *ast.GenDecl:
		return handleGenDecl(src)
	case *ast.FuncDecl:
		return handleFuncDecl(src)
	case *ast.BadDecl:
		return handleBadDecl(src)
	default:
		panic(fmt.Sprintf("Unexpected declaration type: %T", decl))
	}
	return nil
}

func handleBadDecl(decl *ast.BadDecl) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{nodeLayer(decl, badColor, badBackground)}
}

func handleFuncDecl(decl *ast.FuncDecl) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 4)
	if decl.Doc != nil {
		layers = append(layers, nodeLayer(decl.Doc, commentColor))
	}
	layers = append(layers, handleFuncType(decl.Type)...)
	if decl.Recv != nil {
		for _, block := range decl.Recv.List {
			layers = append(layers, nodeLayer(block.Type, typeColor))
		}
	}
	layers = append(layers, nodeLayer(decl.Name, functionColor))
	if decl.Type.Params != nil {
		for _, block := range decl.Type.Params.List {
			layers = append(layers, nodeLayer(block.Type, typeColor))
		}
	}
	if decl.Type.Results != nil {
		for _, block := range decl.Type.Results.List {
			layers = append(layers, nodeLayer(block.Type, typeColor))
		}
	}
	if decl.Body != nil {
		layers = append(layers, handleBlockStmt(decl.Body)...)
	}
	return layers
}

func handleGenDecl(decl *ast.GenDecl) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(decl.Specs)+2)
	if decl.Doc != nil {
		layers = append(layers, nodeLayer(decl.Doc, commentColor))
	}
	var tokColor gxui.Color
	switch {
	case decl.Tok.IsKeyword():
		tokColor = keywordColor
	default:
		panic(fmt.Errorf("Don't know how to handle token %v", decl.Tok))
	}
	layers = append(layers, layer(decl.TokPos, len(decl.Tok.String()), tokColor))
	for _, spec := range decl.Specs {
		layers = append(layers, handleSpec(spec)...)
	}
	return layers
}
