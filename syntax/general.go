// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"go/token"

	"github.com/nelsam/vidar/theme"
)

func (s *Syntax) addFieldList(src *ast.FieldList) {
	if src.Opening != 0 && src.Closing != 0 {
		defer s.rainbowScope(src.Opening, 1, src.Closing, 1)()
	}
	for _, block := range src.List {
		s.addTypeExpr(block.Type)
	}
}

func (s *Syntax) addStructType(src *ast.StructType) {
	s.add(theme.Keyword, src.Struct, len("struct"))
	s.addFieldList(src.Fields)
}

func (s *Syntax) addFuncType(src *ast.FuncType) {
	if src.Func != token.NoPos {
		s.add(theme.Keyword, src.Func, len("func"))
	}
	if src.Params != nil {
		s.addFieldList(src.Params)
	}
	if src.Results != nil {
		s.addFieldList(src.Results)
	}
}

func (s *Syntax) addInterfaceType(src *ast.InterfaceType) {
	s.add(theme.Keyword, src.Interface, len("interface"))
	s.addFieldList(src.Methods)
}

func (s *Syntax) addMapType(src *ast.MapType) {
	s.add(theme.Keyword, src.Map, len("map"))
	s.addTypeExpr(src.Key)
	s.addTypeExpr(src.Value)
}

func (s *Syntax) addArrayType(src *ast.ArrayType) {
	rbrack := src.Lbrack + 1
	if src.Len != nil {
		rbrack = src.Len.End()
		s.addExpr(src.Len)
	}
	defer s.rainbowScope(src.Lbrack, 1, rbrack, 1)()
	s.addTypeExpr(src.Elt)
}

func (s *Syntax) addChanType(src *ast.ChanType) {
	length := len("chan")
	if src.Arrow != token.NoPos {
		length += len("<-")
	}
	s.add(theme.Keyword, src.Begin, length)
	s.addExpr(src.Value)
}

func (s *Syntax) addBasicLit(src *ast.BasicLit) {
	var construct theme.LanguageConstruct
	switch src.Kind {
	case token.INT, token.FLOAT:
		construct = theme.Num
	case token.CHAR, token.STRING:
		construct = theme.String
	default:
		return
	}
	s.addNode(construct, src)
}

func (s *Syntax) addCompositeLit(src *ast.CompositeLit) {
	s.addExpr(src.Type)
	defer s.rainbowScope(src.Lbrace, 1, src.Rbrace, 1)()
	for _, elt := range src.Elts {
		s.addExpr(elt)
	}
}

func (s *Syntax) addCommClause(src *ast.CommClause) {
	length := len("case")
	if src.Comm == nil {
		length = len("default")
	}
	s.add(theme.Keyword, src.Case, length)
	s.addStmt(src.Comm)
	for _, stmt := range src.Body {
		s.addStmt(stmt)
	}
}

func (s *Syntax) addEllipsis(src *ast.Ellipsis) {
	s.add(theme.Keyword, src.Ellipsis, len("..."))
	s.addExpr(src.Elt)
}
