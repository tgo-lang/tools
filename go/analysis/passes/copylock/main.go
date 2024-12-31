// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// The copylock command applies the golang.org/x/tools/go/analysis/passes/copylock
// analysis to the specified packages of Go source code.
package main

import (
	"github.com/tgo-lang/tools/go/analysis/passes/copylock"
	"github.com/tgo-lang/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(copylock.Analyzer) }
