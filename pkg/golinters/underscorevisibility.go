package golinters

import (
	visibility "github.com/dbraley/underscore-visibility-linter"
	"github.com/golangci/golangci-lint/pkg/golinters/goanalysis"
	"golang.org/x/tools/go/analysis"
)

func UnderscoreVisibility() *goanalysis.Linter {

	analyzers := []*analysis.Analyzer{
		visibility.VisibilityAnalyzer,
	}

	return goanalysis.NewLinter(
		"underscorevisibility",
		"checks that symbols starting with _ are only used in the same file",
		analyzers,
		nil,
	).WithLoadMode(goanalysis.LoadModeTypesInfo)
}
