// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"fmt"
	"go/ast"

	"github.com/nelsam/gxui"
)

func handleStmt(stmt ast.Stmt) gxui.CodeSyntaxLayers {
	if stmt == nil {
		return nil
	}
	switch src := stmt.(type) {
	case *ast.BadStmt:
		return handleBadStmt(src)
	case *ast.AssignStmt:
		return handleAssignStmt(src)
	case *ast.CommClause:
		return nil
	case *ast.SwitchStmt:
		return handleSwitchStmt(src)
	case *ast.TypeSwitchStmt:
		return handleTypeSwitchStmt(src)
	case *ast.CaseClause:
		return handleCaseStmt(src)
	case *ast.DeclStmt:
		return handleDecl(src.Decl)
	case *ast.SendStmt:
		return handleSendStmt(src)
	case *ast.SelectStmt:
		return handleSelectStmt(src)
	case *ast.ReturnStmt:
		return handleReturnStmt(src)
	case *ast.RangeStmt:
		return handleRangeStmt(src)
	case *ast.LabeledStmt:
		return handleLabeledStmt(src)
	case *ast.IncDecStmt:
		return handleIncDecStmt(src)
	case *ast.IfStmt:
		return handleIfStmt(src)
	case *ast.GoStmt:
		return handleGoStmt(src)
	case *ast.ForStmt:
		return handleForStmt(src)
	case *ast.ExprStmt:
		return handleExprStmt(src)
	case *ast.EmptyStmt:
		return nil
	case *ast.DeferStmt:
		return handleDeferStmt(src)
	case *ast.BranchStmt:
		return handleBranchStmt(src)
	case *ast.BlockStmt:
		return handleBlockStmt(src)
	default:
		panic(fmt.Errorf("Unknown stmt type: %T", stmt))
	}
}

func handleBadStmt(stmt *ast.BadStmt) gxui.CodeSyntaxLayers {
	return gxui.CodeSyntaxLayers{nodeLayer(stmt, badColor, badBackground)}
}

func handleAssignStmt(stmt *ast.AssignStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, len(stmt.Lhs)+len(stmt.Rhs))
	for _, expr := range stmt.Lhs {
		layers = append(layers, handleExpr(expr)...)
	}
	for _, expr := range stmt.Rhs {
		layers = append(layers, handleExpr(expr)...)
	}
	return layers
}

func handleSwitchStmt(stmt *ast.SwitchStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+3)
	layers = append(layers, layer(stmt.Switch, len("switch"), keywordColor))
	layers = append(layers, handleStmt(stmt.Init)...)
	layers = append(layers, handleExpr(stmt.Tag)...)
	layers = append(layers, handleBlockStmt(stmt.Body)...)
	return layers
}

func handleTypeSwitchStmt(stmt *ast.TypeSwitchStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+3)
	layers = append(layers, layer(stmt.Switch, len("switch"), keywordColor))
	layers = append(layers, handleStmt(stmt.Init)...)
	layers = append(layers, handleStmt(stmt.Assign)...)
	layers = append(layers, handleBlockStmt(stmt.Body)...)
	return layers
}

func handleCaseStmt(stmt *ast.CaseClause) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body)+len(stmt.List))
	length := len("case")
	if stmt.List == nil {
		length = len("default")
	}
	layers = append(layers, layer(stmt.Case, length, keywordColor))
	for _, expr := range stmt.List {
		layers = append(layers, handleExpr(expr)...)
	}
	for _, stmt := range stmt.Body {
		layers = append(layers, handleStmt(stmt)...)
	}
	return layers
}

func handleSendStmt(stmt *ast.SendStmt) gxui.CodeSyntaxLayers {
	return append(handleExpr(stmt.Chan), handleExpr(stmt.Value)...)
}

func handleReturnStmt(stmt *ast.ReturnStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 5)
	layers = append(layers, layer(stmt.Return, len("return"), keywordColor))
	for _, res := range stmt.Results {
		layers = append(layers, handleExpr(res)...)
	}
	return layers
}

func handleRangeStmt(stmt *ast.RangeStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+5)
	layers = append(layers, layer(stmt.For, len("for"), keywordColor))
	layers = append(layers, handleExpr(stmt.X)...)
	layers = append(layers, handleBlockStmt(stmt.Body)...)
	return layers
}

func handleIfStmt(stmt *ast.IfStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+3)
	layers = append(layers, layer(stmt.If, len("if"), keywordColor))
	layers = append(layers, handleStmt(stmt.Init)...)
	layers = append(layers, handleExpr(stmt.Cond)...)
	layers = append(layers, handleBlockStmt(stmt.Body)...)
	layers = append(layers, handleStmt(stmt.Else)...)
	return layers
}

func handleForStmt(stmt *ast.ForStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.Body.List)+5)
	layers = append(layers, layer(stmt.For, len("for"), keywordColor))
	layers = append(layers, handleStmt(stmt.Init)...)
	layers = append(layers, handleExpr(stmt.Cond)...)
	layers = append(layers, handleStmt(stmt.Post)...)
	return layers
}

func handleGoStmt(stmt *ast.GoStmt) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{layer(stmt.Go, len("go"), keywordColor)}
	return append(layers, handleCallExpr(stmt.Call)...)
}

func handleLabeledStmt(stmt *ast.LabeledStmt) gxui.CodeSyntaxLayers {
	// TODO: handle stmt.Label
	return handleStmt(stmt.Stmt)
}

func handleDeferStmt(stmt *ast.DeferStmt) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{layer(stmt.Defer, len("defer"), keywordColor)}
	layers = append(layers, handleCallExpr(stmt.Call)...)
	return layers
}

func handleIncDecStmt(stmt *ast.IncDecStmt) gxui.CodeSyntaxLayers {
	return handleExpr(stmt.X)
}

func handleExprStmt(stmt *ast.ExprStmt) gxui.CodeSyntaxLayers {
	return handleExpr(stmt.X)
}

func handleSelectStmt(stmt *ast.SelectStmt) gxui.CodeSyntaxLayers {
	layers := gxui.CodeSyntaxLayers{layer(stmt.Select, len("select"), keywordColor)}
	layers = append(layers, handleBlockStmt(stmt.Body)...)
	return layers
}

func handleBranchStmt(stmt *ast.BranchStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 2)
	layers = append(layers, layer(stmt.TokPos, len(stmt.Tok.String()), keywordColor))

	// TODO: handle stmt.Label
	return layers
}

func handleBlockStmt(stmt *ast.BlockStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.List)+2)
	layers = append(layers, layer(stmt.Lbrace, 1, defaultRainbow.New()))
	for _, stmt := range stmt.List {
		layers = append(layers, handleStmt(stmt)...)
	}
	return append(layers, layer(stmt.Rbrace, 1, defaultRainbow.Pop()))
}
