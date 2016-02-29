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

func zeroBasedPos(pos token.Pos) uint64 {
	if pos < 1 {
		panic("Positions of 0 are not valid positions")
	}
	return uint64(pos) - 1
}

func nodeLayer(node ast.Node, colors ...gxui.Color) *gxui.CodeSyntaxLayer {
	return layer(node.Pos(), int(node.End()-node.Pos()), colors...)
}

func layer(pos token.Pos, length int, colors ...gxui.Color) *gxui.CodeSyntaxLayer {
	if length == 0 {
		return nil
	}
	if len(colors) == 0 {
		panic("No colors passed to layer()")
	}
	if len(colors) > 2 {
		panic("Only two colors (text and background) are currently supported")
	}
	layer := gxui.CreateCodeSyntaxLayer()
	layer.Add(int(zeroBasedPos(pos)), length)
	layer.SetColor(colors[0])
	if len(colors) > 1 {
		layer.SetBackgroundColor(colors[1])
	}
	return layer
}

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
