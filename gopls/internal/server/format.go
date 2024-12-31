// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"

	"github.com/tgo-lang/tools/gopls/internal/file"
	"github.com/tgo-lang/tools/gopls/internal/golang"
	"github.com/tgo-lang/tools/gopls/internal/label"
	"github.com/tgo-lang/tools/gopls/internal/mod"
	"github.com/tgo-lang/tools/gopls/internal/protocol"
	"github.com/tgo-lang/tools/gopls/internal/work"
	"github.com/tgo-lang/tools/internal/event"
)

func (s *server) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	ctx, done := event.Start(ctx, "lsp.Server.formatting", label.URI.Of(params.TextDocument.URI))
	defer done()

	fh, snapshot, release, err := s.fileOf(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	defer release()

	switch snapshot.FileKind(fh) {
	case file.Mod:
		return mod.Format(ctx, snapshot, fh)
	case file.Go:
		return golang.Format(ctx, snapshot, fh)
	case file.Work:
		return work.Format(ctx, snapshot, fh)
	}
	return nil, nil // empty result
}
