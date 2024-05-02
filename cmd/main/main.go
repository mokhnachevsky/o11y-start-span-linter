package main

import (
	spanchecker "github.com/mokhnachevsky/o11y-start-span-linter"
	"golang.org/x/tools/go/analysis/multichecker"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "./...")
	}

	multichecker.Main(
		spanchecker.SpanChecker,
		spanchecker.RowsCloseChecker,
	)
}
