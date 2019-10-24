package linters

import (
	"golang.org/x/tools/go/analysis"
)

var AllLinters = []*analysis.Analyzer{
	ContextAnalyzer,
}
