package main

import (
	spanchecker "github.com/mokhnachevsky/o11y-start-span-linter"
	"golang.org/x/tools/go/analysis"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{spanchecker.SpanChecker}, nil
}
