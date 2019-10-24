package golinters

import (
	"github.com/dbraley/ctxlint"
	"github.com/golangci/golangci-lint/pkg/golinters/goanalysis"
	"golang.org/x/tools/go/analysis"
)

func CtxLint() *goanalysis.Linter {
	analyzers := []*analysis.Analyzer{
		linters.ContextAnalyzer,
	}

	return goanalysis.NewLinter(
		"ctxlint",
		"prohibits context.Background in most cases, as well as checking "+
			"that Context is always the first parameter and is named ctx.",
		analyzers,
		nil,
	).WithLoadMode(goanalysis.LoadModeTypesInfo)

}
