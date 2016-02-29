// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"log"

	"github.com/nelsam/gxui"
)

func (l layers) handleSpec(spec ast.Spec) gxui.CodeSyntaxLayers {
	switch src := spec.(type) {
	case *ast.ValueSpec:
		return l.handleValueSpec(src)
	case *ast.ImportSpec:
		return l.handleImportSpec(src)
	case *ast.TypeSpec:
		return l.handleTypeSpec(src)
	default:
		log.Printf("Error: Unknown spec type: %T", spec)
		return nil
	}
}

func (l layers) handleValueSpec(val *ast.ValueSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(val.Names)+len(val.Values)+3)
	if val.Doc != nil {
		layers = append(layers, l.nodeLayer(val.Doc, commentColor))
	}
	if val.Comment != nil {
		layers = append(layers, l.nodeLayer(val.Comment, commentColor))
	}
	if val.Type != nil {
		layers = append(layers, l.nodeLayer(val.Type, typeColor))
	}
	return layers
}

func (l layers) handleImportSpec(imp *ast.ImportSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	if imp.Doc != nil {
		layers = append(layers, l.nodeLayer(imp.Doc, commentColor))
	}
	if imp.Comment != nil {
		layers = append(layers, l.nodeLayer(imp.Comment, commentColor))
	}
	// TODO: Decide if there should be more highlighting here.  It
	// seems like the current import highlighting already takes care
	// of it.
	return layers
}

func (l layers) handleTypeSpec(typ *ast.TypeSpec) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	if typ.Doc != nil {
		layers = append(layers, l.nodeLayer(typ.Doc, commentColor))
	}
	if typ.Comment != nil {
		layers = append(layers, l.nodeLayer(typ.Comment, commentColor))
	}
	layers = append(layers, l.handleExpr(typ.Type)...)
	return layers
}
