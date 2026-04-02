package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	multichecker.Main(allAnalyzers()...)
}

func allAnalyzers() []*analysis.Analyzer {
	analyzers := []*analysis.Analyzer{
		// Core analyzers from golang.org/x/tools/go/analysis/passes
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,

		// Custom analyzer.
		ExitCheckAnalyzer,
	}

	// All SA analyzers from staticcheck.
	for _, a := range staticcheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	// Omit redundant nil check on slices, maps, and channels.
	if a := pickAnalyzerByName(simple.Analyzers, "S1009"); a != nil {
		analyzers = append(analyzers, a)
	}
	// Incorrect or missing package comment.
	if a := pickAnalyzerByName(stylecheck.Analyzers, "ST1000"); a != nil {
		analyzers = append(analyzers, a)
	}
	// Convert slice of bytes to string when printing it.
	if a := pickAnalyzerByName(quickfix.Analyzers, "QF1010"); a != nil {
		analyzers = append(analyzers, a)
	}

	return analyzers
}

func pickAnalyzerByName(list []*lint.Analyzer, name string) *analysis.Analyzer {
	for _, item := range list {
		if item != nil && item.Analyzer != nil && item.Analyzer.Name == name {
			return item.Analyzer
		}
	}
	return nil
}
