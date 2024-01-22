package main

import (
	"github.com/mokhnachevsky/o11y-start-span-linter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(spanchecker.SpanChecker)
}
