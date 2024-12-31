// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package analyzer_test

import (
	"testing"

	"github.com/tgo-lang/tools/go/analysis/analysistest"
	inlineanalyzer "github.com/tgo-lang/tools/internal/refactor/inline/analyzer"
)

func TestAnalyzer(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), inlineanalyzer.Analyzer, "a", "b")
}
