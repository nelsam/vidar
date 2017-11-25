// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"log"

	"github.com/nelsam/vidar/theme"
)

func (s *Syntax) addDecl(decl ast.Decl) {
	switch src := decl.(type) {
	case *ast.GenDecl:
		s.addGenDecl(src)
	case *ast.FuncDecl:
		s.addFuncDecl(src)
	case *ast.BadDecl:
		s.addBadDecl(src)
	default:
		log.Printf("Error: Unexpected declaration type: %T", decl)
	}
}

func (s *Syntax) addBadDecl(decl *ast.BadDecl) {
	s.addNode(theme.Bad, decl)
}

func (s *Syntax) addFuncDecl(decl *ast.FuncDecl) {
	if decl.Recv != nil {
		s.addFieldList(decl.Recv)
	}
	s.addNode(theme.Func, decl.Name)
	s.addFuncType(decl.Type)
	if decl.Body != nil {
		s.addBlockStmt(decl.Body)
	}
}

func (s *Syntax) addGenDecl(decl *ast.GenDecl) {
	var tokType theme.LanguageConstruct
	switch {
	case decl.Tok.IsKeyword():
		tokType = theme.Keyword
	default:
		log.Printf("Error: Don't know how to handle token %v", decl.Tok)
		return
	}
	if decl.Lparen != 0 && decl.Rparen != 0 {
		defer s.rainbowScope(decl.Lparen, 1, decl.Rparen, 1)()
	}
	s.add(tokType, decl.TokPos, len(decl.Tok.String()))
	for _, spec := range decl.Specs {
		s.addSpec(spec)
	}
}
