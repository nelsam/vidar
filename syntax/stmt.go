// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package syntax

import (
	"fmt"
	"go/ast"
	"log"

	"github.com/nelsam/gxui"
)

func handleStmt(stmt ast.Stmt) gxui.CodeSyntaxLayers {
	if stmt == nil {
		return nil
	}
	switch src := stmt.(type) {
	case *ast.BadStmt:
		log.Printf("Bad statement: %v", stmt)
		return nil
	case *ast.AssignStmt:
		return nil
	case *ast.SwitchStmt:
		return handleSwitchStmt(src)
	case *ast.TypeSwitchStmt:
		return handleTypeSwitchStmt(src)
	case *ast.CaseClause:
		return handleCaseStmt(src)
	case *ast.SendStmt:
		return nil
	case *ast.SelectStmt:
		return nil
	case *ast.ReturnStmt:
		return handleReturnStmt(src)
	case *ast.RangeStmt:
		return handleRangeStmt(src)
	case *ast.LabeledStmt:
		return nil
	case *ast.IncDecStmt:
		return nil
	case *ast.IfStmt:
		return handleIfStmt(src)
	case *ast.GoStmt:
		return nil
	case *ast.ForStmt:
		return handleForStmt(src)
	case *ast.ExprStmt:
		return handleExprStmt(src)
	case *ast.EmptyStmt:
		return nil
	case *ast.DeferStmt:
		return nil
	case *ast.BranchStmt:
		return handleBranchStmt(src)
	case *ast.BlockStmt:
		return handleBlockStmt(src)
	default:
		panic(fmt.Errorf("Unknown stmt type: %T", stmt))
	}
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

func handleExprStmt(stmt *ast.ExprStmt) gxui.CodeSyntaxLayers {
	return handleExpr(stmt.X)
}

func handleBranchStmt(stmt *ast.BranchStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, 2)
	layers = append(layers, layer(stmt.TokPos, len(stmt.Tok.String()), keywordColor))

	// TODO: handle stmt.Label
	return layers
}

func handleBlockStmt(stmt *ast.BlockStmt) gxui.CodeSyntaxLayers {
	layers := make(gxui.CodeSyntaxLayers, 0, len(stmt.List))
	for _, stmt := range stmt.List {
		layers = append(layers, handleStmt(stmt)...)
	}
	return layers
}
