// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"github.com/codegangsta/cli"
	"github.com/google/gxui"
)

var (
	builtins = []string{}

	background   = gxui.Gray10
	builtinColor = gxui.Color{
		R: 0.9,
		G: 0.3,
		B: 0,
		A: 1.0,
	}
	nilColor = gxui.Color{
		R: 1,
		G: 0.2,
		B: 0.2,
		A: 1.0,
	}
	keywordColor = gxui.Color{
		R: 0,
		G: 0.6,
		B: 0.8,
		A: 1.0,
	}
	functionColor = gxui.Color{
		R: 0.3,
		G: 0.6,
		B: 0,
		A: 1.0,
	}
	typeColor = gxui.Color{
		R: 0.3,
		G: 0.6,
		B: 0.3,
		A: 1.0,
	}
	stringColor = gxui.Color{
		R: 0,
		G: 0.8,
		B: 0,
		A: 1.0,
	}
	commentColor = gxui.Gray60

	app *cli.App
	ctx *cli.Context
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

func layers(filename, text string) (gxui.CodeSyntaxLayers, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, text, 0)
	layers := make(gxui.CodeSyntaxLayers, 0, 100)
	if f.Doc != nil {
		layers = append(layers, nodeLayer(f.Doc, commentColor))
	}
	for _, comment := range f.Comments {
		layers = append(layers, nodeLayer(comment, commentColor))
	}
	for _, decl := range f.Decls {
		var newLayers gxui.CodeSyntaxLayers
		switch src := decl.(type) {
		case *ast.GenDecl:
			newLayers = handleGenDecl(src)
		case *ast.FuncDecl:
			newLayers = handleFuncDecl(src)
		default:
			log.Printf("Unexpected declaration type: %T", decl)
		}
		layers = append(layers, newLayers...)
	}
	for _, unresolved := range f.Unresolved {
		switch unresolved.String() {
		case "append", "cap", "close", "complex", "copy",
			"delete", "imag", "len", "make", "new", "panic",
			"print", "println", "real", "recover":

			layers = append(layers, nodeLayer(unresolved, builtinColor))
		case "nil":
			layers = append(layers, nodeLayer(unresolved, nilColor))
		default:
			log.Printf("Found unresolved declaration %s at position %d", unresolved.String(), unresolved.Pos())
		}
	}
	return layers, err
}

func handleFuncDecl(decl *ast.FuncDecl) gxui.CodeSyntaxLayers {
	layerLen := 3 // doc, func keyword, and function name
	if decl.Recv != nil {
		layerLen += decl.Recv.NumFields()
	}
	if decl.Type.Params != nil {
		layerLen += decl.Type.Params.NumFields()
	}
	if decl.Type.Results != nil {
		layerLen += decl.Type.Results.NumFields()
	}
	layers := make(gxui.CodeSyntaxLayers, 0, layerLen)
	if decl.Doc != nil {
		layers = append(layers, nodeLayer(decl.Doc, commentColor))
	}
	if decl.Type.Func != token.NoPos {
		layers = append(layers, layer(decl.Type.Func, len("func"), keywordColor))
	}
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
		var newLayers gxui.CodeSyntaxLayers
		switch src := spec.(type) {
		case *ast.ValueSpec:
			newLayers = handleValue(src)
		case *ast.ImportSpec:
			newLayers = handleImport(src)
		case *ast.TypeSpec:
			newLayers = handleType(src)
		default:
			panic("Unknown type")
		}
		layers = append(layers, newLayers...)
	}
	return layers
}

func handleValue(val *ast.ValueSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(val.Names)+len(val.Values)+3)
	if val.Doc != nil {
		layers = append(layers, nodeLayer(val.Doc, commentColor))
	}
	if val.Type != nil {
		layers = append(layers, nodeLayer(val.Type, typeColor))
	}
	if val.Comment != nil {
		layers = append(layers, nodeLayer(val.Comment, commentColor))
	}
	return layers
}

func handleImport(*ast.ImportSpec) gxui.CodeSyntaxLayers {
	return nil
}

func handleType(*ast.TypeSpec) gxui.CodeSyntaxLayers {
	return nil
}
