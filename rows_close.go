package spanchecker

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"reflect"
)

var RowsCloseChecker = &analysis.Analyzer{
	Name:     "rowsclosechecker",
	Doc:      "Checks sql rows close",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      RowsCloseCheck,
}

var RowsCloseReport = func(pass *analysis.Pass, tokenPos token.Pos, funcName string) {
	pass.Reportf(tokenPos, "function %s has not closed rows", funcName)
}

func RowsCloseCheck(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok {
				findQueryContextCalls(pass, function)
			}
		}
	}
	return nil, nil
}

func findQueryContextCalls(pass *analysis.Pass, function *ast.FuncDecl) {
	for _, line := range function.Body.List {
		if stmt, ok := line.(*ast.AssignStmt); ok {
			if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
				if callExpr, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
					if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						if selectorExpr.Sel.Name == "QueryContext" {
							if rowsStmt, ok := stmt.Lhs[0].(*ast.Ident); ok {
								if !functionHasDeferredRowsClose(function, rowsStmt.Obj) {
									RowsCloseReport(pass, rowsStmt.Pos(), function.Name.Name)
								}
							}
						}
					}
				}
			}
		}
	}
}

func functionHasDeferredRowsClose(function *ast.FuncDecl, rowsObject *ast.Object) bool {
	objectAssignee, ok := rowsObject.Decl.(*ast.AssignStmt)
	if !ok {
		return false
	}
	objectDeclarationPosition := objectAssignee.TokPos

	for _, line := range function.Body.List {
		if deferStmt, ok := line.(*ast.DeferStmt); ok {
			switch reflect.TypeOf(deferStmt.Call.Fun).String() {
			case "*ast.FuncLit":
				if checkLambdaFunc(deferStmt.Call.Fun.(*ast.FuncLit), objectDeclarationPosition) {
					return true
				}
			case "*ast.SelectorExpr":
				if checkCloseCall(deferStmt.Call.Fun.(*ast.SelectorExpr), objectDeclarationPosition) {
					return true
				}
			}
		}
	}
	return false
}

func checkCloseCall(selExpr *ast.SelectorExpr, objectDeclarationPosition token.Pos) bool {
	if selExpr.Sel.Name != "Close" {
		return false
	}
	if x, ok := selExpr.X.(*ast.Ident); ok {
		if xAssignee, ok := x.Obj.Decl.(*ast.AssignStmt); ok {
			if xAssignee.TokPos == objectDeclarationPosition {
				return true
			}
		}
	}
	return false
}

func checkLambdaFunc(fun *ast.FuncLit, objectDeclarationPosition token.Pos) bool {
	for _, line := range fun.Body.List {
		if stmt, ok := line.(*ast.AssignStmt); ok {
			if len(stmt.Rhs) != 1 {
				continue
			}
			if callExpr, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if checkCloseCall(selExpr, objectDeclarationPosition) {
						return true
					}
				}
			}
		}
	}
	return false
}
