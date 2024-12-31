// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modernize

import (
	"bytes"
	_ "embed"
	"iter"
	"strings"

	"github.com/tgo-lang/lang/ast"
	"github.com/tgo-lang/lang/format"
	"github.com/tgo-lang/lang/token"
	"github.com/tgo-lang/lang/types"

	"github.com/tgo-lang/tools/go/analysis"
	"github.com/tgo-lang/tools/go/analysis/passes/inspect"
	"github.com/tgo-lang/tools/go/ast/inspector"
	"github.com/tgo-lang/tools/gopls/internal/util/astutil"
	"github.com/tgo-lang/tools/internal/analysisinternal"
	"github.com/tgo-lang/tools/internal/astutil/cursor"
	"github.com/tgo-lang/tools/internal/versions"
)

//go:embed doc.go
var doc string

var Analyzer = &analysis.Analyzer{
	Name:     "modernize",
	Doc:      analysisinternal.MustExtractDoc(doc, "modernize"),
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
	URL:      "https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize",
}

func run(pass *analysis.Pass) (any, error) {
	// Decorate pass.Report to suppress diagnostics in generated files.
	//
	// TODO(adonovan): opt: do this more efficiently by interleaving
	// the micro-passes (as described below) and preemptively skipping
	// the entire subtree for each generated *ast.File.
	{
		// Gather information whether file is generated or not.
		generated := make(map[*token.File]bool)
		for _, file := range pass.Files {
			if ast.IsGenerated(file) {
				generated[pass.Fset.File(file.FileStart)] = true
			}
		}
		report := pass.Report
		pass.Report = func(diag analysis.Diagnostic) {
			if _, ok := generated[pass.Fset.File(diag.Pos)]; ok {
				return // skip checking if it's generated code
			}
			report(diag)
		}
	}

	appendclipped(pass)
	bloop(pass)
	efaceany(pass)
	mapsloop(pass)
	minmax(pass)
	sortslice(pass)

	// TODO(adonovan):
	// - more modernizers here; see #70815.
	// - opt: interleave these micro-passes within a single inspection.
	// - solve the "duplicate import" problem (#68765) when a number of
	//   fixes in the same file are applied in parallel and all add
	//   the same import. The tests exhibit the problem.
	// - should all diagnostics be of the form "x can be modernized by y"
	//   or is that a foolish consistency?

	return nil, nil
}

// -- helpers --

// TODO(adonovan): factor with analysisutil.Imports.
func _imports(pkg *types.Package, path string) bool {
	for _, imp := range pkg.Imports() {
		if imp.Path() == path {
			return true
		}
	}
	return false
}

// equalSyntax reports whether x and y are syntactically equal (ignoring comments).
func equalSyntax(x, y ast.Expr) bool {
	sameName := func(x, y *ast.Ident) bool { return x.Name == y.Name }
	return astutil.Equal(x, y, sameName)
}

// formatNode formats n.
func formatNode(fset *token.FileSet, n ast.Node) []byte {
	var buf bytes.Buffer
	format.Node(&buf, fset, n) // ignore errors
	return buf.Bytes()
}

// formatExprs formats a comma-separated list of expressions.
func formatExprs(fset *token.FileSet, exprs []ast.Expr) string {
	var buf strings.Builder
	for i, e := range exprs {
		if i > 0 {
			buf.WriteString(",  ")
		}
		format.Node(&buf, fset, e) // ignore errors
	}
	return buf.String()
}

// isZeroLiteral reports whether e is the literal 0.
func isZeroLiteral(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == "0"
}

// isPackageLevel reports whether obj is the package-level symbol pkg.Name.
func isPackageLevel(obj types.Object, pkgpath, name string) bool {
	pkg := obj.Pkg()
	return pkg != nil &&
		obj.Parent() == pkg.Scope() &&
		obj.Pkg().Path() == pkgpath &&
		obj.Name() == name
}

// filesUsing returns a cursor for each *ast.File in the inspector
// that uses at least the specified version of Go (e.g. "go1.24").
func filesUsing(inspect *inspector.Inspector, info *types.Info, version string) iter.Seq[cursor.Cursor] {
	return func(yield func(cursor.Cursor) bool) {
		for curFile := range cursor.Root(inspect).Children() {
			file := curFile.Node().(*ast.File)
			if !versions.Before(info.FileVersions[file], version) && !yield(curFile) {
				break
			}
		}
	}
}

var (
	builtinAny    = types.Universe.Lookup("any")
	builtinAppend = types.Universe.Lookup("append")
	builtinBool   = types.Universe.Lookup("bool")
	builtinMake   = types.Universe.Lookup("make")
	builtinNil    = types.Universe.Lookup("nil")
)
