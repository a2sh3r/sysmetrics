// Package staticlint contains multichecker for static code analysis.
//
// Multichecker combines several static analyzers:
// - Standard analyzers from golang.org/x/tools/go/analysis/passes
// - SA class analyzers from staticcheck.io
// - Additional analyzers from staticcheck.io
// - Public analyzers
// - Custom exitcheck analyzer
//
// Usage:
//
//	go run cmd/staticlint/main.go ./...
//	go build -o staticlint cmd/staticlint/main.go && ./staticlint ./...
package main

import (
	"encoding/json"
	"go/ast"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Config — configuration file name.
const Config = `config.json`

// ConfigData describes configuration file structure.
type ConfigData struct {
	Staticcheck []string
	Stylecheck  []string
	Simple      []string
	Quickfix    []string
}

// ExitCheckAnalyzer - analyzer that prohibits using os.Exit in main function of main package.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "Prohibits using os.Exit in main function of main package",
	Run:  runExitCheck,
}

// runExitCheck performs code analysis for os.Exit usage in main function.
func runExitCheck(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if callExpr, ok := node.(*ast.CallExpr); ok {
				if isOSExitCall(callExpr) {
					if isInMainFunction(file, node) {
						pass.Reportf(callExpr.Pos(), "prohibited use of os.Exit in main function of main package")
					}
				}
			}
			return true
		})
	}

	return nil, nil
}

// isOSExitCall checks if function call is os.Exit call.
func isOSExitCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "os" && sel.Sel.Name == "Exit"
		}
	}
	return false
}

// isInMainFunction checks if AST node is inside main function.
func isInMainFunction(file *ast.File, node ast.Node) bool {
	var inMain bool

	ast.Inspect(file, func(n ast.Node) bool {
		if n == node {
			return false
		}

		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "main" {
				inMain = true
			} else {
				inMain = false
			}
		}

		return true
	})

	return inMain
}

func main() {
	configPath := Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		appfile, err := os.Executable()
		if err != nil {
			panic(err)
		}
		configPath = filepath.Join(filepath.Dir(appfile), Config)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	mychecks := []*analysis.Analyzer{
		ExitCheckAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		reflectvaluecompare.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	}

	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}

	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	stylechecks := make(map[string]bool)
	for _, v := range cfg.Stylecheck {
		stylechecks[v] = true
	}

	for _, v := range stylecheck.Analyzers {
		if stylechecks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	simplechecks := make(map[string]bool)
	for _, v := range cfg.Simple {
		simplechecks[v] = true
	}

	for _, v := range simple.Analyzers {
		if simplechecks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	quickfixchecks := make(map[string]bool)
	for _, v := range cfg.Quickfix {
		quickfixchecks[v] = true
	}

	for _, v := range quickfix.Analyzers {
		if quickfixchecks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(mychecks...)
}
