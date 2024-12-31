// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// The modernize command suggests (or, with -fix, applies) fixes that
// clarify Go code by using more modern features.
package main

import (
	"github.com/tgo-lang/tools/go/analysis/singlechecker"
	"github.com/tgo-lang/tools/gopls/internal/analysis/modernize"
)

func main() { singlechecker.Main(modernize.Analyzer) }
