package packages

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/mateusz834/tgo/transpiler"
	"github.com/tgo-lang/lang/parser"
	"github.com/tgo-lang/lang/scanner"
	"github.com/tgo-lang/lang/token"
	"github.com/tgo-lang/tools/internal/gocommand"
)

func tgoToGoExt(s string) string {
	return s[:len(s)-len(".tgo")] + ".go"
}

func goToTgoExt(s string) string {
	return s[:len(s)-len(".go")] + ".tgo"
}

type tgoOverlay struct {
	// driverOverlay is a map of file paths to overlaid file contents.
	driverOverlay map[string][]byte

	// fakeGoFiles is a set of files, that were transpiled and the transpiled
	// version is contained in the driverOverlay map.
	// File name always ends with ".go".
	fakeGoFiles map[string]struct{}

	// fakeGoFilesWithImportC is a set of fake go files, that include the "C" package.
	fakeGoFilesWithImportC map[string]struct{}

	// erroneousFakeGoFiles contains every file, that failed parsing or transpilation.
	// Only populated when needsFullyTranspiledSource is true.
	erroneousFakeGoFiles map[string]struct{}

	// needsFullyTranspiledSource, when set to true the fake go files must be transpiled
	// entirely, when set to false, only imports part of the file need to be provided.
	needsFullyTranspiledSource bool
}

// rewriteOverlay transpiled every ".tgo" file in o.driverOverlay into a ".go" file.
func rewriteOverlay(o tgoOverlay) {
	for path, content := range o.driverOverlay {
		if filepath.Ext(path) == ".tgo" {
			asGoFile := tgoToGoExt(path)
			if o.needsFullyTranspiledSource {
				out, err := transpiler.TransileSrc(string(content))
				if err == nil {
					o.driverOverlay[asGoFile] = []byte(out)
				} else {
					o.erroneousFakeGoFiles[asGoFile] = struct{}{}
					src, err := importsOnly(content)
					if err != nil {
						if !errors.Is(err, errImportsC) {
							panic("unreachable")
						}
						o.fakeGoFilesWithImportC[asGoFile] = struct{}{}
					}
					o.driverOverlay[asGoFile] = src
				}
			} else {
				src, err := importsOnly(content)
				if err != nil {
					if !errors.Is(err, errImportsC) {
						panic("unreachable")
					}
					o.fakeGoFilesWithImportC[asGoFile] = struct{}{}
				}
				o.driverOverlay[asGoFile] = src
			}
			o.fakeGoFiles[asGoFile] = struct{}{}
			delete(o.driverOverlay, path)
		}
	}
}

// rewriteFilePatterns rewrittes "file=file.tgo" patterns into "file=file.go" patterns.
// Loads the file from the filesystem, transpiles it and adds it to the overlay (if does not exists).
func rewriteFilePatterns(patterns []string, o tgoOverlay) error {
	for i, pattern := range patterns {
		query, filePath, ok := strings.Cut(pattern, "=")
		if ok && query == "file" && filepath.Ext(filePath) == ".tgo" {
			patterns[i] = tgoToGoExt(pattern)
			if err := addTgoFileToOverlay(o, filePath); err != nil {
				return err
			}
		}
	}
	return nil
}

// The go list driver accepts files as a pattern.
// See: https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns:
//
//	As a special case, if the package list is a list of .go files from a single directory,
//	the command is applied to a single synthesized package made up of exactly those files,
//	ignoring any build constraints in those files and ignoring any other files in the directory.
//
// This function detects such condition and rewrites such patterns (".tgo") to ".go" and populates the overlay.
func rewriteCommandLineArgumentPatterns(patterns []string, o tgoOverlay) error {
	extractPattern := func(pattern string) string {
		query, after, ok := strings.Cut(pattern, "=")
		if !ok {
			return pattern
		} else if query != "file" {
			return after
		}
		return ""
	}

	hasFile := false
outer:
	for _, pattern := range patterns {
		pattern = extractPattern(pattern)
		switch filepath.Ext(pattern) {
		case ".go", ".tgo":
			if filepath.IsAbs(pattern) {
				hasFile = true
				break
			}
			// TODO: same story as in go.dev/cl/680819
			// use cfg.abs(pattern) once we have it avail (merged).
			if _, err := os.Stat(pattern); err == nil {
				hasFile = true
				break outer
			}
		}
	}

	if !hasFile {
		return nil
	}

	for i, pattern := range patterns {
		pattern = extractPattern(pattern)
		if pattern != "" {
			patterns[i] = tgoToGoExt(pattern)
			if err := addTgoFileToOverlay(o, pattern); err != nil {
				return err
			}
		}
	}

	return nil
}

// addTgoFileToOverlay loads the tgoFile from the filesystem, transpiles it
// and adds it to the overlay (if does not exists).
func addTgoFileToOverlay(o tgoOverlay, tgoFile string) error {
	asGoFile := tgoToGoExt(tgoFile)
	if _, ok := o.fakeGoFiles[asGoFile]; ok {
		return nil
	}

	src, err := readTgoOverlayFile(tgoFile, o.needsFullyTranspiledSource)
	if err != nil {
		if errors.Is(err, errImportsC) || errors.Is(err, errTranspilationFailed) {
			if errors.Is(err, errImportsC) {
				o.fakeGoFilesWithImportC[asGoFile] = struct{}{}
			}
			if errors.Is(err, errTranspilationFailed) {
				o.erroneousFakeGoFiles[asGoFile] = struct{}{}
			}
		} else {
			return err
		}
	}

	o.driverOverlay[asGoFile] = src
	o.fakeGoFiles[asGoFile] = struct{}{}
	return nil
}

// fillTgoOverlayBasedOnModules determines all the main modules in the current workspace
// and recursively transpiles every ".tgo" file into ".go" file found in the filesystem and
// puts them into the tgoOverlay.
func fillTgoOverlayBasedOnModules(cfg *Config, patterns []string, o tgoOverlay) error {
	// Skip when running in non-module mode.
	// TODO: test what happens when we are outside of an module and pattern references a module (absolute path??).
	if slices.ContainsFunc(cfg.Env, func(env string) bool {
		env, val, _ := strings.Cut(env, "=")
		return env == "GO111MODULE" && val == "off"
	}) {
		return nil
	}

	// TODO: think, we are using cfg.Overlay here, not tgoOverlay.
	overlayFile, cleanupOverlay, err := gocommand.WriteOverlays(cfg.Overlay)
	if err != nil {
		return err
	}
	defer cleanupOverlay()

	// TODO: now as of https://github.com/golang/go/issues/71075 we should first of all use the "work" pattern, to only
	// transpile the tgo files in the current workspace, but we also need to think about external modules support.
	// We cannot overlay such files (from external modules), so "go list" is going to use the ".go" variants.
	// But while loading them (parsing) we should load the ".tgo" ones. It is fine to load ".tgo" files, assuming
	// that ".go" files are the transpiled variants.
	// We could (at least) verify that the .tgo and .go file variants agree on the imports of the file.

	var r gocommand.Runner
	runGoCmd := func(cfg *Config, dir string, overlay string, verb string, args ...string) (*bytes.Buffer, error) {
		return r.Run(context.Background(), gocommand.Invocation{
			Verb:       verb,
			Args:       args,
			BuildFlags: cfg.BuildFlags,
			ModFile:    cfg.modFile,
			ModFlag:    cfg.modFlag,
			CleanEnv:   cfg.Env != nil,
			Env:        cfg.Env,
			Logf:       cfg.Logf,
			WorkingDir: dir,
			Overlay:    overlay,
		})
	}

	handleDir := func(dir string) error {
		// TODO: use "work" pattern (go.dev/issue/71294) once 1.25 is released.
		b, err := runGoCmd(cfg, dir, overlayFile, "list", "-m", "-json=Dir", "all")
		if err != nil {
			if strings.Contains(err.Error(), "go.mod file not found in current directory or any parent directory") {
				return nil
			}
			return err
		}

		type Module struct {
			Dir string `json:"dir"`
		}

		for d := json.NewDecoder(b); d.More(); {
			var mod Module
			if err := d.Decode(&mod); err != nil {
				return err
			}

			if mod.Dir == "" {
				continue
			}

			// Recursively walk the root directory, and transpile every ".tgo" file into ".go" one.
			err := filepath.WalkDir(mod.Dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && filepath.Ext(path) == ".tgo" {
					if err := addTgoFileToOverlay(o, path); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	if err := handleDir(cfg.Dir); err != nil {
		return err
	}

	handlePattern := func(p string) error {
		if filepath.Ext(p) == "" {
			return nil
		}
		// TODO: abs based on cfg.Dir.
		return handleDir(filepath.Dir(p))
	}

	// file queries might reference files that reside in different module.
	for _, pattern := range patterns {
		if eqidx := strings.Index(pattern, "="); eqidx == -1 {
			if err := handlePattern(pattern); err != nil {
				return err
			}
		} else {
			if err := handlePattern(pattern[eqidx+1:]); err != nil {
				return err
			}
		}
	}

	return nil
}

func fillTgoOverlayBasedOnDriverResponse(dr *DriverResponse, o tgoOverlay) (bool, error) {
	// TODO: might add overlay to module files (which is prohibited 1.25).

	added := false
	for _, pkg := range dr.Packages {
		if pkg.Dir == "" {
			continue
		}
		dir, err := os.Open(pkg.Dir)
		if err != nil {
			return false, err
		}
		files, err := dir.Readdirnames(-1)
		if err != nil {
			return false, err
		}
		for _, tgoFile := range files {
			if filepath.Ext(tgoFile) == ".tgo" {
				tgoFile = filepath.Join(pkg.Dir, tgoFile)
				if err := addTgoFileToOverlay(o, filepath.Join(pkg.Dir, tgoFile)); err != nil {
					return false, err
				}
			}
		}
	}

	return added, nil
}

func rewriteDriverResponse(dr *DriverResponse, o tgoOverlay) {
	rewriteFiles := func(files []string) {
		for i, v := range files {
			if _, ok := o.fakeGoFiles[v]; ok {
				files[i] = goToTgoExt(v)
			}
		}
	}

	rewritePos := func(pos *string) {
		if fileName, after, ok := strings.Cut(*pos, ":"); ok {
			if _, ok := o.fakeGoFiles[fileName]; ok {
				*pos = goToTgoExt(fileName) + ":" + after
			}
		}
	}

	for _, pkg := range dr.Packages {
		for _, file := range pkg.CompiledGoFiles {
			if _, ok := o.erroneousFakeGoFiles[file]; ok {
				// Clear ExportFile and Target, the driver might not have returned any error as we
				// handed to the driver a file with only imports, so the package
				// might not have had any errors, so these fields could be populated.
				pkg.ExportFile = ""
				pkg.Target = ""
				break
			}
		}

		// Reject packages containing ".tgo" files that import "C" (use cgo).
		for _, file := range pkg.GoFiles {
			if _, ok := o.fakeGoFilesWithImportC[file]; ok {
				*pkg = Package{
					ID:      pkg.ID,
					Name:    pkg.Name,
					PkgPath: pkg.PkgPath,
					Dir:     pkg.Dir,
					Errors: []Error{{
						Msg:  fmt.Sprintf(`%v: tgo file cannot import "C"`, goToTgoExt(file)),
						Kind: ListError,
					}},
					Module: pkg.Module,
				}
				break
			}
		}

		for i := range pkg.Errors {
			rewritePos(&pkg.Errors[i].Pos)
		}
		for i := range pkg.depsErrors {
			rewritePos(&pkg.depsErrors[i].Pos)
		}
		rewriteFiles(pkg.GoFiles)
		rewriteFiles(pkg.CompiledGoFiles)
	}
}

func readTgoOverlayFile(path string, needsFullyTranspiledSource bool) ([]byte, error) {
	if needsFullyTranspiledSource {
		return readAndTranspileFile(path)
	}
	return readFileImports(path)
}

var errTranspilationFailed = errors.New("failed to transpile file")

func readAndTranspileFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	out, err := transpiler.TransileSrc(string(src))
	if err != nil {
		out, err2 := importsOnly(src)
		return out, errors.Join(errTranspilationFailed, err2)
	}

	return []byte(out), nil
}

var pool8KBytes = sync.Pool{
	New: func() any {
		return &[8192]byte{}
	},
}

// readFileImports reads a file from the filesystem and returns the imports portion of the file.
func readFileImports(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read only the first 8KiB of the file, we only need the imports of the file,
	// so don't bother reading the entire file, we will fallback later to reading
	// the entire file (if needed).
	// TODO: maybe even 4KiB would be enough? Check at common repos, to see.
	buf := pool8KBytes.Get().(*[8192]byte)
	defer pool8KBytes.Put(buf)

	n, err := io.ReadFull(f, buf[:])
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	if src, err := importsOnlyPartialFile(buf[:n], err == io.ErrUnexpectedEOF); len(src) != 0 || err != nil {
		return bytes.Clone(src), err
	}

	// Imports might not have fit in the first 8KiB, or the file is
	// invalid (has syntax errors), so fallback reading the entire file
	// and using go/parser to get proper error recovery.

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return importsOnly(src)
}

var errImportsC = errors.New("file imports C")

// importsOnly returns only the imports part of the source code.
func importsOnly(c []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", c, parser.ImportsOnly|parser.SkipObjectResolution|parser.AllErrors)
	if end := f.End(); end.IsValid() {
		var err error
		for _, i := range f.Imports {
			i, err := strconv.Unquote(i.Path.Value)
			if err == nil && i == "C" {
				err = errImportsC
			}
		}
		return c[:fset.File(end).Offset(end)], err
	}
	return c, nil // some invalid file (it is even missing a package directive)
}

// importsOnlyPartialFile returns the imports part of the source code.
// The input src can be a partial source of a file.
// Returns nil, when there could be more imports (if the entire file was read instead).
func importsOnlyPartialFile(src []byte, entireFile bool) (out []byte, err error) {
	file := token.NewFileSet().AddFile("", -1, len(src))

	errorHandlerCalled := false
	eh := func(pos token.Position, msg string) { errorHandlerCalled = true }
	defer func() {
		if errorHandlerCalled {
			out = nil
		}
	}()

	var s scanner.Scanner
	s.Init(file, src, eh, 0)

	_, tok, _ := s.Scan()
	if tok != token.PACKAGE {
		return nil, nil
	}

	_, tok, lit := s.Scan()
	if tok != token.IDENT {
		return nil, nil
	}

	scanImport := func() bool {
		switch tok {
		case token.IDENT, token.PERIOD:
			_, tok, lit = s.Scan()
		case token.STRING:
		default:
			return false
		}

		ok := tok == token.STRING
		var i string
		i, err = strconv.Unquote(lit)
		if err == nil && i == "C" {
			err = errImportsC
		}

		_, tok, lit = s.Scan()
		return ok
	}

	for {
		var pos token.Pos
		pos, tok, lit = s.Scan()
		switch tok {
		case token.IMPORT:
			_, tok, lit = s.Scan()
			switch tok {
			case token.LPAREN:
				for tok != token.RPAREN {
					if !scanImport() {
						return nil, err
					}
				}
			default:
				if !scanImport() {
					return nil, err
				}
			}
		case token.FUNC, token.CONST, token.VAR, token.TYPE:
			// TODO: do not include comments between last import and these tokens.
			return src[:file.Offset(pos)], err
		case token.EOF:
			if entireFile {
				// TODO: do not include comments between last import and these tokens.
				return src[:file.Offset(pos)], err
			}
			return nil, err // there could be more import statements
		default:
			return nil, err // unexpected token
		}
	}
}

//func rewriteDriverResponse(cfg *Config, dr *DriverResponse, o tgoOverlay, external bool) {
//	// To the packages driver, we might pass invalid files (imports only part of a file),
//	// if we get a package, that includes such file, then we need to handle it specially.
//	for _, pkg := range dr.Packages {
//		var errs []error
//		for _, file := range pkg.CompiledGoFiles {
//			if err, ok := o.erroneousFakeGoFiles[file]; ok {
//				errs = append(errs, err)
//			}
//		}
//		if len(errs) != 0 {
//			// TODO: with https://go-review.googlesource.com/c/tools/+/682855 this might not be really
//			// needed. As we will remove the error anyway.
//			pkg.Errors = nil
//			for _, err := range errs {
//				pkg.Errors = append(pkg.Errors, Error{
//					Kind: ListError,
//					Msg:  err.Error(), // TODO: split errors.
//				})
//			}
//
//			// Clear ExportFile and Target, the driver might not have returned any error as we
//			// handed to the driver a file with only imports, so the package
//			// might not have had any errors, so these fields could be populated.
//			pkg.ExportFile = ""
//			pkg.Target = ""
//		}
//	}
//
//	rewriteFiles := func(files []string) {
//		for i, v := range files {
//			if _, ok := o.fakeGoFiles[v]; ok {
//				files[i] = goToTgoExt(v)
//			}
//		}
//	}
//
//	rewritePos := func(pos *string) {
//		if fileName, after, ok := strings.Cut(*pos, ":"); ok {
//			if _, ok := o.fakeGoFiles[fileName]; ok {
//				*pos = goToTgoExt(fileName) + ":" + after
//			}
//		}
//	}
//
//	// TODO: also same thing for "Target" field.
//
//	for _, pkg := range dr.Packages {
//		// TODO: we can also in case of (usesExportData(cfg) || external) transpile the file instead.
//		// Or do that dynamically, we only need export data when NeedExportFile is set, otherwise we can
//		// do whatever we want (what about errors??).
//		// Or if someone needs export data, then we can produce in??
//		// But first figure out whether it is safe to pass transpiled file (and) fset handling of the export file.
//		//
//		// Also to avoid  work we can parse the file and do not include function bodies (name returns, end add return stmt).
//		// This would work, but the expoort data also contains other data:
//		// The export data files produced by the compiler contain additional details related to generics, inlining,
//		// and other optimizations that cannot be decoded by the Read function.
//
//		// TODO: use export from compiler and see what is donen with the fset.
//
//		// TODO: i feel like we should transpile it AND clear the export data
//		// or do a simple transpile and clear ExportData (only globals), keep funcs
//		// clear?
//		if len(pkg.CompiledGoFiles) != 0 && (len(pkg.Errors) != 0 || pkg.ExportFile != "") && (external || usesExportData(cfg)) {
//			for _, v := range pkg.GoFiles {
//				if _, ok := o.fakeGoFiles[v]; ok {
//					// We are not "transpiling" ".tgo" files, tgo to go conversion only
//					// includes imports, so ExportFile might be invalid and it might contain
//					// Errors (unused imports, undefined globals (from other files)).
//					// We will get correct errors, after refine.
//					if pkg.ExportFile == "" {
//						pkg.Errors = nil
//					}
//					pkg.ExportFile = ""
//
//					// TODO: file load error (like permission, non-syntax related), we might
//					// miss an error on an unused global function that was failed to read?
//					// TODO: import cycle?
//					break
//				}
//			}
//		}
//
//		for i := range pkg.Errors {
//			rewritePos(&pkg.Errors[i].Pos)
//		}
//		for i := range pkg.depsErrors {
//			rewritePos(&pkg.depsErrors[i].Pos)
//		}
//		rewriteFiles(pkg.GoFiles)
//		rewriteFiles(pkg.CompiledGoFiles)
//	}
//}
//
