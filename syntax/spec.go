// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"fmt"
	"go/ast"

	"github.com/nelsam/gxui"
)

func handleSpec(spec ast.Spec) gxui.CodeSyntaxLayers {
	switch src := spec.(type) {
	case *ast.ValueSpec:
		return handleValueSpec(src)
	case *ast.ImportSpec:
		return handleImportSpec(src)
	case *ast.TypeSpec:
		return handleTypeSpec(src)
	default:
		panic(fmt.Errorf("Unknown spec type: %T", spec))
	}
}

func handleValueSpec(val *ast.ValueSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(val.Names)+len(val.Values)+3)
	if val.Doc != nil {
		layers = append(layers, nodeLayer(val.Doc, commentColor))
	}
	if val.Comment != nil {
		layers = append(layers, nodeLayer(val.Comment, commentColor))
	}
	if val.Type != nil {
		layers = append(layers, nodeLayer(val.Type, typeColor))
	}
	return layers
}

func handleImportSpec(imp *ast.ImportSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	if imp.Doc != nil {
		layers = append(layers, nodeLayer(imp.Doc, commentColor))
	}
	if imp.Comment != nil {
		layers = append(layers, nodeLayer(imp.Comment, commentColor))
	}
	// TODO: Decide if there should be more highlighting here.  It
	// seems like the current import highlighting already takes care
	// of it.
	return layers
}

func handleTypeSpec(typ *ast.TypeSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	if typ.Doc != nil {
		layers = append(layers, nodeLayer(typ.Doc, commentColor))
	}
	if typ.Comment != nil {
		layers = append(layers, nodeLayer(typ.Comment, commentColor))
	}
	layers = append(layers, handleExpr(typ.Type)...)
	return layers
}
