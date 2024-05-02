package spanchecker

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var RowsCloseChecker = &analysis.Analyzer{
	Name:     "rowsclosechecker",
	Doc:      "Checks sql rows close",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      RowsCloseCheck,
}

var PassReport2 = func(pass *analysis.Pass, tokenPos token.Pos, funcName string) {
	pass.Reportf(tokenPos, "function %s has not closed rows", funcName)
}

func RowsCloseCheck(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok {
				CheckFunction(pass, function)
			}
		}
	}
	return nil, nil
}

func CheckFunction(pass *analysis.Pass, function *ast.FuncDecl) {
	for _, line := range function.Body.List {
		if stmt, ok := line.(*ast.AssignStmt); ok {
			if len(stmt.Rhs) == 1 {
				if callExpr, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
					selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					if selExpr.Sel.Name != "QueryContext" {
						continue
					}
					if len(stmt.Lhs) != 2 {
						continue
					}
					rowsStmt, ok := stmt.Lhs[0].(*ast.Ident)
					if !ok {
						continue
					}
					if !findDeferredClose(function, rowsStmt.Obj) {
						PassReport2(pass, selExpr.X.Pos(), function.Name.Name)
					}
				}
			}
		}
	}
}

func findDeferredClose(function *ast.FuncDecl, object *ast.Object) bool {
	objectAssignee, ok := object.Decl.(*ast.AssignStmt)
	if !ok {
		return false
	}
	objectDeclarationPosition := objectAssignee.TokPos

	for _, line := range function.Body.List {
		if stmt, ok := line.(*ast.DeferStmt); ok {
			selExpr, ok := stmt.Call.Fun.(*ast.SelectorExpr)
			if !ok {
				funcLit, ok := stmt.Call.Fun.(*ast.FuncLit)
				if !ok {
					continue
				}
				if checkLambdaDefer(funcLit, objectDeclarationPosition) {
					return true
				} else {
					continue
				}
			}
			if checkCloseCall(selExpr, objectDeclarationPosition) {
				return true
			} else {
				continue
			}
		}
	}
	return false
}

func checkCloseCall(selExpr *ast.SelectorExpr, objectDeclarationPosition token.Pos) bool {
	if selExpr.Sel.Name != "Close" {
		return false
	}
	x, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	xAssignee, ok := x.Obj.Decl.(*ast.AssignStmt)
	if !ok {
		return false
	}
	if xAssignee.TokPos == objectDeclarationPosition {
		return true
	}
	return false
}

func checkLambdaDefer(funcLit *ast.FuncLit, objectDeclarationPosition token.Pos) bool {
	for _, line := range funcLit.Body.List {
		if stmt, ok := line.(*ast.AssignStmt); ok {
			if len(stmt.Rhs) != 1 {
				continue
			}
			if callExpr, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if checkCloseCall(selExpr, objectDeclarationPosition) {
						return true
					} else {
						continue
					}
				}
			}
		}
	}
	return false
}
