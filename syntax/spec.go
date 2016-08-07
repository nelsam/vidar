// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"log"
)

func (s *Syntax) addSpec(spec ast.Spec) {
	switch src := spec.(type) {
	case *ast.ValueSpec:
		s.addValueSpec(src)
	case *ast.ImportSpec:
		s.addImportSpec(src)
	case *ast.TypeSpec:
		s.addTypeSpec(src)
	default:
		log.Printf("Error: Unknown spec type: %T", spec)
	}
}

func (s *Syntax) addValueSpec(val *ast.ValueSpec) {
	if val.Type != nil {
		s.addNode(s.Theme.Colors.Type, val.Type)
	}
	for _, expr := range val.Values {
		s.addExpr(expr)
	}
}

func (s *Syntax) addImportSpec(imp *ast.ImportSpec) {
	// TODO: Decide if there should be highlighting here.  It
	// seems like the current import and comment highlighting
	// already takes care of it.
}

func (s *Syntax) addTypeSpec(typ *ast.TypeSpec) {
	s.addExpr(typ.Type)
}
