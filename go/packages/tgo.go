package packages

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tgo-lang/lang/parser"
	"github.com/tgo-lang/lang/token"
	"github.com/tgo-lang/tools/internal/gocommand"
)

func tgoToGoExt(s string) string {
	return s[:len(s)-len(".tgo")] + ".go"
}

func goToTgoExt(s string) string {
	return s[:len(s)-len(".go")] + ".tgo"
}

func tgoToGoImportsOnly(c []byte) []byte {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", c, parser.SkipObjectResolution|parser.ImportsOnly|parser.AllErrors)
	if end := f.End(); end.IsValid() {
		return c[:fset.File(end).Offset(end)]
	}
	return c // some invalid file (it is even missing a package directive)
}

type tgoOverlay struct {
	driverOverlay map[string][]byte
	addedGoFiles  map[string]struct{}
}

func rewriteOverlay(o tgoOverlay) {
	for path, content := range o.driverOverlay {
		if filepath.Ext(path) == ".tgo" {
			asGoFile := tgoToGoExt(path)
			o.driverOverlay[asGoFile] = tgoToGoImportsOnly(content)
			o.addedGoFiles[asGoFile] = struct{}{}
			delete(o.driverOverlay, path)
		}
	}
}

func addTgoFileToOverlay(o tgoOverlay, tgoFile string) error {
	asGoFile := tgoToGoExt(tgoFile)
	if _, ok := o.addedGoFiles[asGoFile]; ok {
		return nil
	}

	tgoFileContents, err := os.ReadFile(tgoFile)
	if err != nil {
		return err
	}

	o.driverOverlay[asGoFile] = tgoToGoImportsOnly(tgoFileContents)
	o.addedGoFiles[asGoFile] = struct{}{}

	return nil
}

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

func runGoCmd(r *gocommand.Runner, cfg *Config, overlay string, verb string, args ...string) (*bytes.Buffer, error) {
	return r.Run(context.Background(), gocommand.Invocation{
		Verb:       verb,
		Args:       args,
		BuildFlags: cfg.BuildFlags,
		ModFile:    cfg.modFile,
		ModFlag:    cfg.modFlag,
		CleanEnv:   cfg.Env != nil,
		Env:        cfg.Env,
		Logf:       cfg.Logf,
		WorkingDir: cfg.Dir,
		Overlay:    overlay,
	})
}

func fillTgoOverlayBasedOnModules(cfg *Config, o tgoOverlay) error {
	overlayFile, cleanupOverlay, err := gocommand.WriteOverlays(cfg.Overlay)
	if err != nil {
		return err
	}
	defer cleanupOverlay()

	var r gocommand.Runner
	b, err := runGoCmd(&r, cfg, overlayFile, "list", "-m", "-json=Dir", "all")
	if err != nil {
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

func fillTgoOverlayBasedOnDriverResponse(dr *DriverResponse, o tgoOverlay) (bool, error) {
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

func rewriteDriverResponse(cfg *Config, dr *DriverResponse, o tgoOverlay, external bool) {
	rewriteFiles := func(files []string) {
		for i, v := range files {
			if _, ok := o.addedGoFiles[v]; ok {
				files[i] = goToTgoExt(v)
			}
		}
	}

	rewritePos := func(pos *string) {
		if fileName, after, ok := strings.Cut(*pos, ":"); ok {
			if _, ok := o.addedGoFiles[fileName]; ok {
				*pos = goToTgoExt(fileName) + ":" + after
			}
		}
	}

	// TODO: also same thing for "Target" field.

	for _, pkg := range dr.Packages {
		// TODO: we can also in case of (usesExportData(cfg) || external) transpile the file instead.
		// Or do that dynamically, we only need export data when NeedExportFile is set, otherwise we can
		// do whatever we want (what about errors??).
		// Or if someone needs export data, then we can produce in??
		// But first figure out whether it is safe to pass transpiled file (and) fset handling of the export file.
		//
		// Also to avoid  work we can parse the file and do not include function bodies (name returns, end add return stmt).
		// This would work, but the expoort data also contains other data:
		// The export data files produced by the compiler contain additional details related to generics, inlining,
		// and other optimizations that cannot be decoded by the Read function.

		// TODO: use export from compiler and see what is donen with the fset.

		// TODO: i feel like we should transpile it AND clear the export data
		// or do a simple transpile and clear ExportData (only globals), keep funcs
		// clear?
		if len(pkg.CompiledGoFiles) != 0 && (len(pkg.Errors) != 0 || pkg.ExportFile != "") && (external || usesExportData(cfg)) {
			for _, v := range pkg.GoFiles {
				if _, ok := o.addedGoFiles[v]; ok {
					// We are not "transpiling" ".tgo" files, tgo to go conversion only
					// includes imports, so ExportFile might be invalid and it might contain
					// Errors (unused imports, undefined globals (from other files)).
					// We will get correct errors, after refine.
					if pkg.ExportFile == "" {
						pkg.Errors = nil
					}
					pkg.ExportFile = ""

					// TODO: file load error (like permission, non-syntax related), we might
					// miss an error on an unused global function that was failed to read?
					// TODO: import cycle?
					break
				}
			}
		}

		// TODO: in case of an overlay they are wrong?
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
