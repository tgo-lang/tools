// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"

	"github.com/tgo-lang/lang/parser"
	"github.com/tgo-lang/lang/token"

	"github.com/tgo-lang/tools/gopls/internal/cache/metadata"
	"github.com/tgo-lang/tools/gopls/internal/cache/parsego"
	"github.com/tgo-lang/tools/gopls/internal/cache/symbols"
	"github.com/tgo-lang/tools/gopls/internal/file"
	"github.com/tgo-lang/tools/gopls/internal/filecache"
	"github.com/tgo-lang/tools/gopls/internal/protocol"
	"github.com/tgo-lang/tools/gopls/internal/util/bug"
	"github.com/tgo-lang/tools/internal/event"
	"golang.org/x/sync/errgroup"
)

// Symbols extracts and returns symbol information for every file contained in
// a loaded package. It awaits snapshot loading.
//
// If workspaceOnly is set, this only includes symbols from files in a
// workspace package. Otherwise, it returns symbols from all loaded packages.
func (s *Snapshot) Symbols(ctx context.Context, ids ...PackageID) ([]*symbols.Package, error) {
	meta := s.MetadataGraph()

	res := make([]*symbols.Package, len(ids))
	var g errgroup.Group
	g.SetLimit(runtime.GOMAXPROCS(-1)) // symbolizing is cpu bound
	for i, id := range ids {
		g.Go(func() error {
			mp := meta.Packages[id]
			if mp == nil {
				return bug.Errorf("missing metadata for %q", id)
			}

			key, fhs, err := symbolKey(ctx, mp, s)
			if err != nil {
				return err
			}

			if data, err := filecache.Get(symbolsKind, key); err == nil {
				res[i] = symbols.Decode(data)
				return nil
			} else if err != filecache.ErrNotFound {
				bug.Reportf("internal error reading symbol data: %v", err)
			}

			pgfs, err := s.view.parseCache.parseFiles(ctx, token.NewFileSet(), parsego.Full&^parser.ParseComments, false, fhs...)
			if err != nil {
				return err
			}
			pkg := symbols.New(pgfs)

			// Store the resulting data in the cache.
			go func() {
				data := pkg.Encode()
				if err := filecache.Set(symbolsKind, key, data); err != nil {
					event.Error(ctx, fmt.Sprintf("storing symbol data for %s", id), err)
				}
			}()

			res[i] = pkg
			return nil
		})
	}

	return res, g.Wait()
}

func symbolKey(ctx context.Context, mp *metadata.Package, fs file.Source) (file.Hash, []file.Handle, error) {
	seen := make(map[protocol.DocumentURI]bool)
	var fhs []file.Handle
	for _, list := range [][]protocol.DocumentURI{mp.GoFiles, mp.CompiledGoFiles} {
		for _, uri := range list {
			if !seen[uri] {
				seen[uri] = true
				fh, err := fs.ReadFile(ctx, uri)
				if err != nil {
					return file.Hash{}, nil, err // context cancelled
				}
				fhs = append(fhs, fh)
			}
		}
	}

	hasher := sha256.New()
	fmt.Fprintf(hasher, "symbols: %s\n", mp.PkgPath)
	fmt.Fprintf(hasher, "files: %d\n", len(fhs))
	for _, fh := range fhs {
		fmt.Fprintln(hasher, fh.Identity())
	}
	var hash file.Hash
	hasher.Sum(hash[:0])
	return hash, fhs, nil
}
