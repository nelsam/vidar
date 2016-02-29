// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"github.com/nelsam/gxui"
)

type layers struct {
	syntaxLayers *gxui.CodeSyntaxLayers
	fileSet      *token.FileSet
}

func (l layers) nodeLayer(node ast.Node, colors ...gxui.Color) *gxui.CodeSyntaxLayer {
	return l.layer(node.Pos(), int(node.End()-node.Pos()), colors...)
}

func (l layers) append(newSyntaxLayers ...*gxui.CodeSyntaxLayer) {
	*l.syntaxLayers = append(*l.syntaxLayers, newSyntaxLayers...)
}

func (l layers) layer(pos token.Pos, length int, colors ...gxui.Color) *gxui.CodeSyntaxLayer {
	if length == 0 {
		return nil
	}
	if len(colors) == 0 {
		log.Printf("Error: No colors passed to layer()")
		return nil
	}
	if len(colors) > 2 {
		err := errors.New("Only two colors (text and background) are currently supported")
		log.Printf("Error: Expected %d to be <= 2: %s", len(colors), err)
	}
	layer := gxui.CreateCodeSyntaxLayer()
	layer.Add(int(l.fileSet.Position(pos).Offset), length)
	layer.SetColor(colors[0])
	if len(colors) > 1 {
		layer.SetBackgroundColor(colors[1])
	}
	return layer
}

func Layers(filename, text string) (gxui.CodeSyntaxLayers, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, text, 0)
	syntaxLayers := make(gxui.CodeSyntaxLayers, 0, 100)
	layers := layers{
		syntaxLayers: &syntaxLayers,
		fileSet:      fset,
	}
	if f.Doc != nil {
		layers.append(layers.nodeLayer(f.Doc, commentColor))
	}
	if f.Package.IsValid() {
		layers.append(layers.layer(f.Package, len("package"), keywordColor))
	}
	for _, importSpec := range f.Imports {
		layers.append(layers.nodeLayer(importSpec, stringColor))
	}
	for _, comment := range f.Comments {
		layers.append(layers.nodeLayer(comment, commentColor))
	}
	for _, decl := range f.Decls {
		layers.append(layers.handleDecl(decl)...)
	}
	for _, unresolved := range f.Unresolved {
		layers.append(layers.handleUnresolved(unresolved)...)
	}
	return *layers.syntaxLayers, err
}
