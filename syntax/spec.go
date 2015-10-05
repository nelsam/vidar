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
	if val.Type != nil {
		layers = append(layers, nodeLayer(val.Type, typeColor))
	}
	if val.Comment != nil {
		layers = append(layers, nodeLayer(val.Comment, commentColor))
	}
	return layers
}

func handleImportSpec(*ast.ImportSpec) gxui.CodeSyntaxLayers {
	return nil
}

func handleTypeSpec(*ast.TypeSpec) gxui.CodeSyntaxLayers {
	return nil
}
