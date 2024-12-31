// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modernize_test

import (
	"testing"

	"github.com/tgo-lang/tools/go/analysis/analysistest"
	"github.com/tgo-lang/tools/gopls/internal/analysis/modernize"
)

func Test(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), modernize.Analyzer,
		"appendclipped",
		"bloop",
		"efaceany",
		"mapsloop",
		"minmax",
		"sortslice",
	)
}
