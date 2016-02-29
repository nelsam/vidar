// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"log"

	"github.com/nelsam/gxui"
)

func (l layers) handleStmt(stmt ast.Stmt) gxui.CodeSyntaxLayers {
	if stmt == nil {
		return nil
	}
	switch src := stmt.(type) {
	case *ast.BadStmt:
		return l.handleBadStmt(src)
	case *ast.AssignStmt:
		return l.handleAssignStmt(src)
	case *ast.CommClause:
		return nil
	case *ast.SwitchStmt:
		return l.handleSwitchStmt(src)
	case *ast.TypeSwitchStmt:
		return l.handleTypeSwitchStmt(src)
	case *ast.CaseClause:
		return l.handleCaseStmt(src)
	case *ast.DeclStmt:
		return l.handleDecl(src.Decl)
	case *ast.SendStmt:
		return l.handleSendStmt(src)
	case *ast.SelectStmt:
		return l.handleSelectStmt(src)
	case *ast.ReturnStmt:
		return l.handleReturnStmt(src)
	case *ast.RangeStmt:
		return l.handleRangeStmt(src)
	case *ast.LabeledStmt:
		return l.handleLabeledStmt(src)
	case *ast.IncDecStmt:
		return l.handleIncDecStmt(src)
	case *ast.IfStmt:
		return l.handleIfStmt(src)
	case *ast.GoStmt:
		return l.handleGoStmt(src)
	case *ast.ForStmt:
		return l.handleForStmt(src)
	case *ast.ExprStmt:
		return l.handleExprStmt(src)
	case *ast.EmptyStmt:
		return nil
	case *ast.DeferStmt:
		return l.handleDeferStmt(src)
	case *ast.BranchStmt:
		return l.handleBranchStmt(src)
	case *ast.BlockStmt:
		return l.handleBlockStmt(src)
	default:
		log.Printf("Error: Unknown stmt type: %T", stmt)
		return nil
	}
}

func (l layers) handleBadStmt(stmt *ast.BadStmt) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{l.nodeLayer(stmt, badColor, badBackground)}
}

func (l layers) handleAssignStmt(stmt *ast.AssignStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, len(stmt.Lhs)+len(stmt.Rhs))
	for _, expr := range stmt.Lhs {
		layers = append(layers, l.handleExpr(expr)...)
	}
	for _, expr := range stmt.Rhs {
		layers = append(layers, l.handleExpr(expr)...)
	}
	return layers
}

func (l layers) handleSwitchStmt(stmt *ast.SwitchStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+3)
	layers = append(layers, l.layer(stmt.Switch, len("switch"), keywordColor))
	layers = append(layers, l.handleStmt(stmt.Init)...)
	layers = append(layers, l.handleExpr(stmt.Tag)...)
	layers = append(layers, l.handleBlockStmt(stmt.Body)...)
	return layers
}

func (l layers) handleTypeSwitchStmt(stmt *ast.TypeSwitchStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+3)
	layers = append(layers, l.layer(stmt.Switch, len("switch"), keywordColor))
	layers = append(layers, l.handleStmt(stmt.Init)...)
	layers = append(layers, l.handleStmt(stmt.Assign)...)
	layers = append(layers, l.handleBlockStmt(stmt.Body)...)
	return layers
}

func (l layers) handleCaseStmt(stmt *ast.CaseClause) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body)+len(stmt.List))
	length := len("case")
	if stmt.List == nil {
		length = len("default")
	}
	layers = append(layers, l.layer(stmt.Case, length, keywordColor))
	for _, expr := range stmt.List {
		layers = append(layers, l.handleExpr(expr)...)
	}
	for _, stmt := range stmt.Body {
		layers = append(layers, l.handleStmt(stmt)...)
	}
	return layers
}

func (l layers) handleSendStmt(stmt *ast.SendStmt) gxui.CodeSyntaxLayers {
	return append(l.handleExpr(stmt.Chan), l.handleExpr(stmt.Value)...)
}

func (l layers) handleReturnStmt(stmt *ast.ReturnStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	layers = append(layers, l.layer(stmt.Return, len("return"), keywordColor))
	for _, res := range stmt.Results {
		layers = append(layers, l.handleExpr(res)...)
	}
	return layers
}

func (l layers) handleRangeStmt(stmt *ast.RangeStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+5)
	layers = append(layers, l.layer(stmt.For, len("for"), keywordColor))
	layers = append(layers, l.handleExpr(stmt.X)...)
	layers = append(layers, l.handleBlockStmt(stmt.Body)...)
	return layers
}

func (l layers) handleIfStmt(stmt *ast.IfStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+3)
	layers = append(layers, l.layer(stmt.If, len("if"), keywordColor))
	layers = append(layers, l.handleStmt(stmt.Init)...)
	layers = append(layers, l.handleExpr(stmt.Cond)...)
	layers = append(layers, l.handleBlockStmt(stmt.Body)...)
	layers = append(layers, l.handleStmt(stmt.Else)...)
	return layers
}

func (l layers) handleForStmt(stmt *ast.ForStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+5)
	layers = append(layers, l.layer(stmt.For, len("for"), keywordColor))
	layers = append(layers, l.handleStmt(stmt.Init)...)
	layers = append(layers, l.handleExpr(stmt.Cond)...)
	layers = append(layers, l.handleStmt(stmt.Post)...)
	return layers
}

func (l layers) handleGoStmt(stmt *ast.GoStmt) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{l.layer(stmt.Go, len("go"), keywordColor)}
	return append(layers, l.handleCallExpr(stmt.Call)...)
}

func (l layers) handleLabeledStmt(stmt *ast.LabeledStmt) gxui.CodeSyntaxLayers {
	// TODO: handle stmt.Label
	return l.handleStmt(stmt.Stmt)
}

func (l layers) handleDeferStmt(stmt *ast.DeferStmt) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{l.layer(stmt.Defer, len("defer"), keywordColor)}
	layers = append(layers, l.handleCallExpr(stmt.Call)...)
	return layers
}

func (l layers) handleIncDecStmt(stmt *ast.IncDecStmt) gxui.CodeSyntaxLayers {
	return l.handleExpr(stmt.X)
}

func (l layers) handleExprStmt(stmt *ast.ExprStmt) gxui.CodeSyntaxLayers {
	return l.handleExpr(stmt.X)
}

func (l layers) handleSelectStmt(stmt *ast.SelectStmt) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{l.layer(stmt.Select, len("select"), keywordColor)}
	layers = append(layers, l.handleBlockStmt(stmt.Body)...)
	return layers
}

func (l layers) handleBranchStmt(stmt *ast.BranchStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 2)
	layers = append(layers, l.layer(stmt.TokPos, len(stmt.Tok.String()), keywordColor))

	// TODO: handle stmt.Label
	return layers
}

func (l layers) handleBlockStmt(stmt *ast.BlockStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.List)+2)
	layers = append(layers, l.layer(stmt.Lbrace, 1, defaultRainbow.New()))
	for _, stmt := range stmt.List {
		layers = append(layers, l.handleStmt(stmt)...)
	}
	return append(layers, l.layer(stmt.Rbrace, 1, defaultRainbow.Pop()))
}
