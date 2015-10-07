// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/nelsam/gxui"
)

func nodeLayer(node ast.Node, color gxui.Color) *gxui.CodeSyntaxLayer {
	return layer(node.Pos(), int(node.End()-node.Pos()), color)
}

func layer(pos token.Pos, length int, color gxui.Color) *gxui.CodeSyntaxLayer {
	layer := gxui.CreateCodeSyntaxLayer()
	layer.Add(int(pos-1), length)
	layer.SetColor(color)
	return layer
}

// TODO: highlight matching parenthesis/braces/brackets
func Layers(filename, text string) (gxui.CodeSyntaxLayers, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, text, 0)
	layers := make(gxui.CodeSyntaxLayers, 0, 100)
	if f.Doc != nil {
		layers = append(layers, nodeLayer(f.Doc, commentColor))
	}
	if f.Package.IsValid() {
		layers = append(layers, layer(f.Package, len("package"), keywordColor))
	}
	for _, importSpec := range f.Imports {
		layers = append(layers, nodeLayer(importSpec, stringColor))
	}
	for _, comment := range f.Comments {
		layers = append(layers, nodeLayer(comment, commentColor))
	}
	for _, decl := range f.Decls {
		layers = append(layers, handleDecl(decl)...)
	}
	for _, unresolved := range f.Unresolved {
		layers = append(layers, handleUnresolved(unresolved)...)
	}
	return layers, err
}
