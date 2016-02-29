// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"fmt"
	"go/ast"

	"github.com/nelsam/gxui"
)

func (l layers) handleDecl(decl ast.Decl) gxui.CodeSyntaxLayers {
	switch src := decl.(type) {
	case *ast.GenDecl:
		return l.handleGenDecl(src)
	case *ast.FuncDecl:
		return l.handleFuncDecl(src)
	case *ast.BadDecl:
		return l.handleBadDecl(src)
	default:
		panic(fmt.Sprintf("Unexpected declaration type: %T", decl))
	}
	return nil
}

func (l layers) handleBadDecl(decl *ast.BadDecl) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{l.nodeLayer(decl, badColor, badBackground)}
}

func (l layers) handleFuncDecl(decl *ast.FuncDecl) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 4)
	if decl.Doc != nil {
		layers = append(layers, l.nodeLayer(decl.Doc, commentColor))
	}
	layers = append(layers, l.handleFuncType(decl.Type)...)
	if decl.Recv != nil {
		layers = append(layers, l.handleFieldList(decl.Recv)...)
	}
	layers = append(layers, l.nodeLayer(decl.Name, functionColor))
	layers = append(layers, l.handleFuncType(decl.Type)...)
	if decl.Body != nil {
		layers = append(layers, l.handleBlockStmt(decl.Body)...)
	}
	return layers
}

func (l layers) handleGenDecl(decl *ast.GenDecl) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(decl.Specs)+2)
	if decl.Doc != nil {
		layers = append(layers, l.nodeLayer(decl.Doc, commentColor))
	}
	var tokColor gxui.Color
	switch {
	case decl.Tok.IsKeyword():
		tokColor = keywordColor
	default:
		panic(fmt.Errorf("Don't know how to handle token %v", decl.Tok))
	}
	if decl.Lparen != 0 {
		layers = append(layers, l.layer(decl.Lparen, 1, defaultRainbow.New()))
	}
	layers = append(layers, l.layer(decl.TokPos, len(decl.Tok.String()), tokColor))
	for _, spec := range decl.Specs {
		layers = append(layers, l.handleSpec(spec)...)
	}
	if decl.Rparen != 0 {
		layers = append(layers, l.layer(decl.Rparen, 1, defaultRainbow.Pop()))
	}
	return layers
}
