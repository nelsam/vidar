// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"go/token"
)

func (s *Syntax) addFieldList(src *ast.FieldList) {
	if src.Opening != 0 {
		s.add(defaultRainbow.New(), src.Opening, 1)
	}
	for _, block := range src.List {
		s.addNode(s.Theme.Colors.Type, block.Type)
	}
	if src.Closing != 0 {
		s.add(defaultRainbow.Pop(), src.Closing, 1)
	}
}

func (s *Syntax) addStructType(src *ast.StructType) {
	s.add(s.Theme.Colors.Keyword, src.Struct, len("struct"))
	s.addFieldList(src.Fields)
}

func (s *Syntax) addFuncType(src *ast.FuncType) {
	if src.Func != token.NoPos {
		s.add(s.Theme.Colors.Keyword, src.Func, len("func"))
	}
	if src.Params != nil {
		s.addFieldList(src.Params)
	}
	if src.Results != nil {
		s.addFieldList(src.Results)
	}
}

func (s *Syntax) addInterfaceType(src *ast.InterfaceType) {
	s.add(s.Theme.Colors.Keyword, src.Interface, len("interface"))
	s.addFieldList(src.Methods)
}

func (s *Syntax) addMapType(src *ast.MapType) {
	s.add(s.Theme.Colors.Keyword, src.Map, len("map"))
	s.addExpr(src.Key)
	s.addExpr(src.Value)
}

func (s *Syntax) addArrayType(src *ast.ArrayType) {
	s.add(defaultRainbow.New(), src.Lbrack, 1)
	s.add(defaultRainbow.Pop(), src.Lbrack+1, 1)
	s.addExpr(src.Len)
	s.addExpr(src.Elt)
}

func (s *Syntax) addBasicLit(src *ast.BasicLit) {
	var color Color
	switch src.Kind {
	case token.INT, token.FLOAT:
		color = s.Theme.Colors.Num
	case token.CHAR, token.STRING:
		color = s.Theme.Colors.String
	default:
		return
	}
	s.addNode(color, src)
}

func (s *Syntax) addCompositeLit(src *ast.CompositeLit) {
	s.addExpr(src.Type)
	s.add(defaultRainbow.New(), src.Lbrace, 1)
	for _, elt := range src.Elts {
		s.addExpr(elt)
	}
	s.add(defaultRainbow.Pop(), src.Rbrace, 1)
}

func (s *Syntax) addCommClause(src *ast.CommClause) {
	length := len("case")
	if src.Comm == nil {
		length = len("default")
	}
	s.add(s.Theme.Colors.Keyword, src.Case, length)
	s.addStmt(src.Comm)
	for _, stmt := range src.Body {
		s.addStmt(stmt)
	}
}
