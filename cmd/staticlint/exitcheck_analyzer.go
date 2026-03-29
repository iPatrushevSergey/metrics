package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer reports os.Exit and log.Fatal / Fatalf / Fatalln everywhere
// except inside func main of package main (the program entrypoint).
// Other packages and all other functions in package main are checked.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "forbid os.Exit and log.Fatal* except inside func main of package main",
	Run:  runExitCheck,
}

const msgForbidden = "forbidden: os.Exit or log.Fatal outside func main of package main"

func runExitCheck(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name == nil || d.Body == nil {
					continue
				}
				if pass.Pkg.Name() == "main" && d.Recv == nil && d.Name.Name == "main" {
					continue
				}
				inspectForbidden(pass, d.Body)
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
						inspectForbidden(pass, val)
					}
				}
			}
		}
	}

	return nil, nil
}

func inspectForbidden(pass *analysis.Pass, root ast.Node) {
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
		if isOSExitCall(pass, call) || isLogFatalCall(pass, call) {
			pass.Reportf(call.Pos(), "%s", msgForbidden)
		}
		return true
	})
}

func skipGoBuild(pass *analysis.Pass, pos token.Pos) bool {
	fileName := filepath.ToSlash(pass.Fset.Position(pos).Filename)
	return strings.Contains(fileName, "/go-build/")
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
