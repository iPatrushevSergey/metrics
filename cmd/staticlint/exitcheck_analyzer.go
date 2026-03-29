package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer reports:
//   - os.Exit and log.Fatal / Fatalf / Fatalln everywhere except func main of package main;
//   - the built-in panic everywhere, including func main (no exceptions).
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "forbid os.Exit and log.Fatal* except in func main of package main; forbid built-in panic everywhere",
	Run:  runExitCheck,
}

const (
	msgForbiddenExitFatal = "forbidden: os.Exit or log.Fatal outside func main of package main"
	msgForbiddenPanic     = "forbidden: built-in panic"
)

func runExitCheck(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name == nil || d.Body == nil {
					continue
				}
				exemptExitFatal := pass.Pkg.Name() == "main" && d.Recv == nil && d.Name.Name == "main"
				inspectCalls(pass, d.Body, !exemptExitFatal)
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, val := range vs.Values {
						inspectCalls(pass, val, true)
					}
				}
			}
		}
	}

	return nil, nil
}

// inspectCalls checks call sites under root. If checkExitFatal is false, os.Exit and
// log.Fatal* are skipped (only func main of package main passes false). panic is always checked.
func inspectCalls(pass *analysis.Pass, root ast.Node, checkExitFatal bool) {
	if root == nil {
		return
	}
	ast.Inspect(root, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if skipGoBuild(pass, call.Pos()) {
			return true
		}
		if isBuiltinPanic(pass, call) {
			pass.Reportf(call.Pos(), "%s", msgForbiddenPanic)
		}
		if checkExitFatal && (isOSExitCall(pass, call) || isLogFatalCall(pass, call)) {
			pass.Reportf(call.Pos(), "%s", msgForbiddenExitFatal)
		}
		return true
	})
}

func skipGoBuild(pass *analysis.Pass, pos token.Pos) bool {
	fileName := filepath.ToSlash(pass.Fset.Position(pos).Filename)
	return strings.Contains(fileName, "/go-build/")
}

func isBuiltinPanic(pass *analysis.Pass, call *ast.CallExpr) bool {
	id, ok := call.Fun.(*ast.Ident)
	if !ok || id.Name != "panic" {
		return false
	}
	obj, ok := pass.TypesInfo.Uses[id]
	if !ok {
		return false
	}
	b, ok := obj.(*types.Builtin)
	return ok && b.Name() == "panic"
}

func isOSExitCall(pass *analysis.Pass, call *ast.CallExpr) bool {
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

func isLogFatalCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil {
		return false
	}
	switch sel.Sel.Name {
	case "Fatal", "Fatalf", "Fatalln":
	default:
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

	return pkgName.Imported().Path() == "log"
}
