// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The unusedresult command applies the golang.org/x/tools/go/analysis/passes/unusedresult
// analysis to the specified packages of Go source code.
package main

import (
	"github.com/tgo-lang/tools/go/analysis/passes/unusedresult"
	"github.com/tgo-lang/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(unusedresult.Analyzer) }
