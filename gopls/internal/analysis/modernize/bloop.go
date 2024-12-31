// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modernize

import (
	"fmt"

	"github.com/tgo-lang/lang/ast"
	"github.com/tgo-lang/lang/token"
	"github.com/tgo-lang/lang/types"

	"github.com/tgo-lang/tools/go/analysis"
	"github.com/tgo-lang/tools/go/analysis/passes/inspect"
	"github.com/tgo-lang/tools/go/ast/inspector"
	"github.com/tgo-lang/tools/go/types/typeutil"
	"github.com/tgo-lang/tools/internal/astutil/cursor"
	"github.com/tgo-lang/tools/internal/typesinternal"
)

// bloop updates benchmarks that use "for range b.N", replacing it
// with go1.24's b.Loop() and eliminating any preceding
// b.{Start,Stop,Reset}Timer calls.
//
// Variants:
//
//	for i := 0; i < b.N; i++ {}  =>   for b.Loop() {}
//	for range b.N {}
func bloop(pass *analysis.Pass) {
	if !_imports(pass.Pkg, "testing") {
		return
	}

	info := pass.TypesInfo

	// edits computes the text edits for a matched for/range loop
	// at the specified cursor. b is the *testing.B value, and
	// (start, end) is the portion using b.N to delete.
	edits := func(cur cursor.Cursor, b ast.Expr, start, end token.Pos) (edits []analysis.TextEdit) {
		// Within the same function, delete all calls to
		// b.{Start,Stop,Timer} that precede the loop.
		filter := []ast.Node{(*ast.ExprStmt)(nil), (*ast.FuncLit)(nil)}
		fn, _ := enclosingFunc(cur)
		fn.Inspect(filter, func(cur cursor.Cursor, push bool) (descend bool) {
			if push {
				node := cur.Node()
				if is[*ast.FuncLit](node) {
					return false // don't descend into FuncLits (e.g. sub-benchmarks)
				}
				stmt := node.(*ast.ExprStmt)
				if stmt.Pos() > start {
					return false // not preceding: stop
				}
				if call, ok := stmt.X.(*ast.CallExpr); ok {
					fn := typeutil.StaticCallee(info, call)
					if fn != nil &&
						(isMethod(fn, "testing", "B", "StopTimer") ||
							isMethod(fn, "testing", "B", "StartTimer") ||
							isMethod(fn, "testing", "B", "ResetTimer")) {

						// Delete call statement.
						// TODO(adonovan): delete following newline, or
						// up to start of next stmt? (May delete a comment.)
						edits = append(edits, analysis.TextEdit{
							Pos: stmt.Pos(),
							End: stmt.End(),
						})
					}
				}
			}
			return true
		})

		// Replace ...b.N... with b.Loop().
		return append(edits, analysis.TextEdit{
			Pos:     start,
			End:     end,
			NewText: fmt.Appendf(nil, "%s.Loop()", formatNode(pass.Fset, b)),
		})
	}

	// Find all for/range statements.
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	loops := []ast.Node{
		(*ast.ForStmt)(nil),
		(*ast.RangeStmt)(nil),
	}
	for curFile := range filesUsing(inspect, info, "go1.24") {
		for curLoop := range curFile.Preorder(loops...) {
			switch n := curLoop.Node().(type) {
			case *ast.ForStmt:
				// for _; i < b.N; _ {}
				if cmp, ok := n.Cond.(*ast.BinaryExpr); ok && cmp.Op == token.LSS {
					if sel, ok := cmp.Y.(*ast.SelectorExpr); ok &&
						sel.Sel.Name == "N" &&
						isPtrToNamed(info.TypeOf(sel.X), "testing", "B") {

						delStart, delEnd := n.Cond.Pos(), n.Cond.End()

						// Eliminate variable i if no longer needed:
						//  for i := 0; i < b.N; i++ {
						//    ...no references to i...
						//  }
						body, _ := curLoop.LastChild()
						if assign, ok := n.Init.(*ast.AssignStmt); ok &&
							assign.Tok == token.DEFINE &&
							len(assign.Rhs) == 1 &&
							isZeroLiteral(assign.Rhs[0]) &&
							is[*ast.IncDecStmt](n.Post) &&
							n.Post.(*ast.IncDecStmt).Tok == token.INC &&
							equalSyntax(n.Post.(*ast.IncDecStmt).X, assign.Lhs[0]) &&
							!uses(info, body, info.Defs[assign.Lhs[0].(*ast.Ident)]) {

							delStart, delEnd = n.Init.Pos(), n.Post.End()
						}

						pass.Report(analysis.Diagnostic{
							// Highlight "i < b.N".
							Pos:      n.Cond.Pos(),
							End:      n.Cond.End(),
							Category: "bloop",
							Message:  "b.N can be modernized using b.Loop()",
							SuggestedFixes: []analysis.SuggestedFix{{
								Message:   "Replace b.N with b.Loop()",
								TextEdits: edits(curLoop, sel.X, delStart, delEnd),
							}},
						})
					}
				}

			case *ast.RangeStmt:
				// for range b.N {} -> for b.Loop() {}
				//
				// TODO(adonovan): handle "for i := range b.N".
				if sel, ok := n.X.(*ast.SelectorExpr); ok &&
					n.Key == nil &&
					n.Value == nil &&
					sel.Sel.Name == "N" &&
					isPtrToNamed(info.TypeOf(sel.X), "testing", "B") {

					pass.Report(analysis.Diagnostic{
						// Highlight "range b.N".
						Pos:      n.Range,
						End:      n.X.End(),
						Category: "bloop",
						Message:  "b.N can be modernized using b.Loop()",
						SuggestedFixes: []analysis.SuggestedFix{{
							Message:   "Replace b.N with b.Loop()",
							TextEdits: edits(curLoop, sel.X, n.Range, n.X.End()),
						}},
					})
				}
			}
		}
	}
}

// isPtrToNamed reports whether t is type "*pkgpath.Name".
func isPtrToNamed(t types.Type, pkgpath, name string) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		named, ok := ptr.Elem().(*types.Named)
		return ok &&
			named.Obj().Name() == name &&
			named.Obj().Pkg().Path() == pkgpath
	}
	return false
}

// uses reports whether the subtree cur contains a use of obj.
func uses(info *types.Info, cur cursor.Cursor, obj types.Object) bool {
	for curId := range cur.Preorder((*ast.Ident)(nil)) {
		if info.Uses[curId.Node().(*ast.Ident)] == obj {
			return true
		}
	}
	return false
}

// isMethod reports whether fn is pkgpath.(T).Name.
func isMethod(fn *types.Func, pkgpath, T, name string) bool {
	if recv := fn.Signature().Recv(); recv != nil {
		_, recvName := typesinternal.ReceiverNamed(recv)
		return recvName != nil &&
			isPackageLevel(recvName.Obj(), pkgpath, T) &&
			fn.Name() == name
	}
	return false
}

// enclosingFunc returns the cursor for the innermost Func{Decl,Lit}
// that encloses (or is) c, if any.
//
// TODO(adonovan): consider adding:
//
//	func (Cursor) AnyEnclosing(filter ...ast.Node) (Cursor bool)
//	func (Cursor) Enclosing[N ast.Node]() (Cursor, bool)
//
// See comments at [cursor.Cursor.Stack].
func enclosingFunc(c cursor.Cursor) (cursor.Cursor, bool) {
	for {
		switch c.Node().(type) {
		case *ast.FuncLit, *ast.FuncDecl:
			return c, true
		case nil:
			return cursor.Cursor{}, false
		}
		c = c.Parent()
	}
}
