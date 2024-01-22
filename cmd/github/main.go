package main

import (
	"fmt"
	spanchecker "github.com/mokhnachevsky/o11y-start-span-linter"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "./...")
	}
	spanchecker.PassReport = func(pass *analysis.Pass, lit *ast.BasicLit, expectedSpanName string) {
		pos := pass.Fset.Position(lit.ValuePos)
		fmt.Printf("::error file=%s,line=%d::%s\n", pos.Filename, pos.Line, fmt.Sprintf("bad span name %s (expected: %s)", lit.Value, expectedSpanName))
	}

	singlechecker.Main(spanchecker.SpanChecker)
}
