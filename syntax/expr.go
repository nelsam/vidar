// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"log"
)

func (s *Syntax) addTypeExpr(expr ast.Expr) {
	orig := s.Theme.Colors.Ident
	s.Theme.Colors.Ident = s.Theme.Colors.Type
	s.addExpr(expr)
	s.Theme.Colors.Ident = orig
}

func (s *Syntax) addExpr(expr ast.Expr) {
	if expr == nil {
		return
	}
	switch src := expr.(type) {
	case *ast.ArrayType:
		s.addArrayType(src)
	case *ast.BadExpr:
		s.addBadExpr(src)
	case *ast.BasicLit:
		s.addBasicLit(src)
	case *ast.BinaryExpr:
		s.addBinaryExpr(src)
	case *ast.CallExpr:
		s.addCallExpr(src)
	case *ast.ChanType:
		s.addChanType(src)
	case *ast.FuncLit:
		s.addFuncLitExpr(src)
	case *ast.FuncType:
		s.addFuncType(src)
	case *ast.IndexExpr:
		s.addIndexExpr(src)
	case *ast.InterfaceType:
		s.addInterfaceType(src)
	case *ast.KeyValueExpr:
		s.addKeyValueExpr(src)
	case *ast.MapType:
		s.addMapType(src)
	case *ast.ParenExpr:
		s.addParenExpr(src)
	case *ast.SelectorExpr:
		s.addSelectorExpr(src)
	case *ast.SliceExpr:
		s.addSliceExpr(src)
	case *ast.StructType:
		s.addStructType(src)
	case *ast.StarExpr:
		s.addStarExpr(src)
	case *ast.TypeAssertExpr:
		s.addTypeAssertExpr(src)
	case *ast.UnaryExpr:
		s.addUnaryExpr(src)
	case *ast.Ellipsis:
		s.addEllipsis(src)
	case *ast.Ident:
		switch src.Name {
		case "append", "cap", "close", "complex", "copy",
			"delete", "imag", "len", "make", "new", "panic",
			"print", "println", "real", "recover":

			s.addNode(s.Theme.Colors.Builtin, src)
		case "nil":
			s.addNode(s.Theme.Colors.Nil, src)
		default:
			s.addNode(s.Theme.Colors.Ident, src)
		}
	case *ast.CompositeLit:
		s.addCompositeLit(src)
	default:
		log.Printf("Error: Unknown expression type: %T", expr)
	}
}

func (s *Syntax) addBadExpr(src *ast.BadExpr) {
	s.addNode(s.Theme.Colors.Bad, src)
}

func (s *Syntax) addBinaryExpr(src *ast.BinaryExpr) {
	s.addExpr(src.X)
	s.addExpr(src.Y)
}

func (s *Syntax) addCallExpr(src *ast.CallExpr) {
	s.addExpr(src.Fun)
	s.add(s.Theme.Rainbow.New(), src.Lparen, 1)
	for _, arg := range src.Args {
		s.addExpr(arg)
	}
	s.add(s.Theme.Rainbow.Pop(), src.Rparen, 1)
}

func (s *Syntax) addFuncLitExpr(src *ast.FuncLit) {
	s.addFuncType(src.Type)
	s.addBlockStmt(src.Body)
}

func (s *Syntax) addIndexExpr(src *ast.IndexExpr) {
	s.addExpr(src.X)
	s.add(s.Theme.Rainbow.New(), src.Lbrack, 1)
	s.addExpr(src.Index)
	s.add(s.Theme.Rainbow.Pop(), src.Rbrack, 1)
}

func (s *Syntax) addKeyValueExpr(src *ast.KeyValueExpr) {
	s.addExpr(src.Key)
	s.addExpr(src.Value)
}

func (s *Syntax) addParenExpr(src *ast.ParenExpr) {
	s.add(s.Theme.Rainbow.New(), src.Lparen, 1)
	s.addExpr(src.X)
	s.add(s.Theme.Rainbow.Pop(), src.Rparen, 1)
}

func (s *Syntax) addSelectorExpr(src *ast.SelectorExpr) {
	s.addExpr(src.X)
	s.addNode(s.Theme.Colors.Func, src.Sel)
}

func (s *Syntax) addSliceExpr(src *ast.SliceExpr) {
	s.addExpr(src.X)
	s.add(s.Theme.Rainbow.New(), src.Lbrack, 1)
	s.addExpr(src.Low)
	s.addExpr(src.High)
	s.addExpr(src.Max)
	s.add(s.Theme.Rainbow.Pop(), src.Rbrack, 1)
}

func (s *Syntax) addStarExpr(src *ast.StarExpr) {
	s.addExpr(src.X)
}

func (s *Syntax) addTypeAssertExpr(src *ast.TypeAssertExpr) {
	s.addExpr(src.X)
	s.add(s.Theme.Rainbow.New(), src.Lparen, 1)
	s.addExpr(src.Type)
	s.add(s.Theme.Rainbow.Pop(), src.Rparen, 1)
}

func (s *Syntax) addUnaryExpr(src *ast.UnaryExpr) {
	s.addExpr(src.X)
}
