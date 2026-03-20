package main

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer reports direct os.Exit calls inside func main in package main.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "forbid direct os.Exit calls in package main func main",
	Run:  runExitCheck,
}

func runExitCheck(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if isDirectOSExitCall(pass, call) {
					fileName := filepath.ToSlash(pass.Fset.Position(call.Pos()).Filename)
					if strings.Contains(fileName, "/go-build/") {
						return true
					}
					pass.Reportf(call.Pos(), "direct os.Exit call in main is forbidden")
				}
				return true
			})
		}
	}

	return nil, nil
}

func isDirectOSExitCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Exit" {
		return false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	obj := pass.TypesInfo.Uses[pkgIdent]
	pkgName, ok := obj.(*types.PkgName)
	if !ok || pkgName.Imported() == nil {
		return false
	}

	return pkgName.Imported().Path() == "os"
}
