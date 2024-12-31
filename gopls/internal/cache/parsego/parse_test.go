// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parsego_test

import (
	"context"
	"testing"

	"github.com/tgo-lang/lang/ast"
	"github.com/tgo-lang/lang/token"

	"github.com/tgo-lang/tools/gopls/internal/cache/parsego"
	"github.com/tgo-lang/tools/gopls/internal/util/safetoken"
	"github.com/tgo-lang/tools/internal/tokeninternal"
)

// TODO(golang/go#64335): we should have many more tests for fixed syntax.

func TestFixPosition_Issue64488(t *testing.T) {
	// This test reproduces the conditions of golang/go#64488, where a type error
	// on fixed syntax overflows the token.File.
	const src = `
package foo

func _() {
	type myThing struct{}
	var foo []myThing
	for ${1:}, ${2:} := range foo {
	$0
}
}
`

	pgf, _ := parsego.Parse(context.Background(), token.NewFileSet(), "file://foo.go", []byte(src), parsego.Full, false)
	fset := tokeninternal.FileSetFor(pgf.Tok)
	ast.Inspect(pgf.File, func(n ast.Node) bool {
		if n != nil {
			posn := safetoken.StartPosition(fset, n.Pos())
			if !posn.IsValid() {
				t.Fatalf("invalid position for %T (%v): %v not in [%d, %d]", n, n, n.Pos(), pgf.Tok.Base(), pgf.Tok.Base()+pgf.Tok.Size())
			}
		}
		return true
	})
}
