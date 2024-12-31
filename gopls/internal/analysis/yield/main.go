// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// The yield command applies the yield analyzer to the specified
// packages of Go source code.
package main

import (
	"github.com/tgo-lang/tools/go/analysis/singlechecker"
	"github.com/tgo-lang/tools/gopls/internal/analysis/yield"
)

func main() { singlechecker.Main(yield.Analyzer) }
