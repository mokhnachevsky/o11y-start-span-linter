package spanchecker

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var SpanChecker = &analysis.Analyzer{
	Name:     "spanchecker",
	Doc:      "Checks o11y StartSpan function calls",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var PassReport = func(pass *analysis.Pass, lit *ast.BasicLit, expectedSpanName string) {
	pass.Reportf(lit.ValuePos, "bad span name %s (expected: %s)", lit.Value, expectedSpanName)
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok {
				expectedSpanName := fmt.Sprintf(`"%s"`, genExpectedSpanName(pass, function))

				for _, l := range function.Body.List {
					if stmt, ok := l.(*ast.AssignStmt); ok {
						if lit, ok := extractSpanName(stmt); ok {
							if lit.Value != expectedSpanName {
								PassReport(pass, lit, expectedSpanName)
							}
						}
					}
				}

			}

		}
	}
	return nil, nil
}

func genExpectedSpanName(pass *analysis.Pass, function *ast.FuncDecl) string {
	if function.Recv == nil {
		// Regular function
		return fmt.Sprintf("%s.%s", pass.Pkg.Name(), function.Name.Name)
	} else {
		// Method
		for _, recv := range function.Recv.List {
			if expr, ok := recv.Type.(*ast.StarExpr); ok {
				if ident, ok := expr.X.(*ast.Ident); ok {
					return fmt.Sprintf("(*%s).%s", ident.Name, function.Name.Name)
				}
			}
		}
	}

	return function.Name.Name
}

func extractSpanName(stmt *ast.AssignStmt) (*ast.BasicLit, bool) {
	if len(stmt.Rhs) == 1 {
		if callExpr, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
			if len(callExpr.Args) != 2 {
				return nil, false
			}
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if selExpr.Sel.Name != "StartSpan" {
					return nil, false
				}
			}

			if lit, ok := callExpr.Args[1].(*ast.BasicLit); ok {
				return lit, true
			}
		}
	}

	return nil, false
}
