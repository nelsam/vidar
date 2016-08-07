// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"log"
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
	s.addNode(s.Theme.Colors.Bad, decl)
}

func (s *Syntax) addFuncDecl(decl *ast.FuncDecl) {
	s.addFuncType(decl.Type)
	if decl.Recv != nil {
		s.addFieldList(decl.Recv)
	}
	s.addNode(s.Theme.Colors.Func, decl.Name)
	s.addFuncType(decl.Type)
	if decl.Body != nil {
		s.addBlockStmt(decl.Body)
	}
}

func (s *Syntax) addGenDecl(decl *ast.GenDecl) {
	var tokColor Color
	switch {
	case decl.Tok.IsKeyword():
		tokColor = s.Theme.Colors.Keyword
	default:
		log.Printf("Error: Don't know how to handle token %v", decl.Tok)
		return
	}
	if decl.Lparen != 0 {
		s.add(defaultRainbow.New(), decl.Lparen, 1)
	}
	s.add(tokColor, decl.TokPos, len(decl.Tok.String()))
	for _, spec := range decl.Specs {
		s.addSpec(spec)
	}
	if decl.Rparen != 0 {
		s.add(defaultRainbow.Pop(), decl.Rparen, 1)
	}
}
