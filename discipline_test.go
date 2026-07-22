package sty

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSourceDiscipline(t *testing.T) {
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != "." && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.StructType:
				for _, field := range n.Fields.List {
					if isContextSelector(field.Type, "Context") {
						t.Errorf("%s: context.Context must not be stored in a struct field", fset.Position(field.Pos()))
					}
				}
			case *ast.CallExpr:
				if isContextSelector(n.Fun, "WithValue") {
					t.Errorf("%s: context.WithValue is forbidden", fset.Position(n.Pos()))
				}
			case *ast.ImportSpec:
				if isStagePath(path) {
					importPath, err := strconv.Unquote(n.Path.Value)
					if err == nil && importPath == "github.com/robkerr1992/sty" {
						t.Errorf("%s: package stage must not import sty", fset.Position(n.Pos()))
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func isContextSelector(expr ast.Expr, name string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != name {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "context"
}

func isStagePath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return strings.HasPrefix(clean, "stage/")
}
