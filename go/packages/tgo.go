package packages

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tgo-lang/lang/parser"
	"github.com/tgo-lang/lang/token"
	"github.com/tgo-lang/tools/internal/gocommand"
)

func tgoToGoExt(s string) string {
	return s[:len(s)-len(".tgo")] + ".go"
}

func tgoToGo(c []byte) []byte {
	// TODO: think about this (https://github.com/golang/go/issues/70725)
	f, _ := parser.ParseFile(token.NewFileSet(), "", c, parser.SkipObjectResolution|parser.ImportsOnly)
	if len(f.Decls) != 0 {
		//lastImportSpec := f.Decls[len(f.Decls)-1].(*ast.GenDecl) // TODO: can this be a BadDecl?
		lastImportSpec := f.Decls[len(f.Decls)-1]
		return c[:lastImportSpec.End()]
	} else if f.Name != nil {
		return c[:f.Name.End()]
	} else {
		// TODO: Misisng AllErrors, or invalid file (no package directive).
		return c
	}
}

func fillTgoOverlayBasedOnModules(cfg *Config, driverOverlay map[string][]byte, addedGoFiles map[string]struct{}) error {
	var r gocommand.Runner
	b, err := r.Run(context.Background(), gocommand.Invocation{
		Verb:       "list",
		Args:       []string{"-m", "-json=Dir", "all"},
		Env:        cfg.Env,
		WorkingDir: cfg.Dir,
	})
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
				asGoFile := tgoToGoExt(path)
				if _, ok := driverOverlay[asGoFile]; ok {
					return nil
				}

				tgoFileContents, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				driverOverlay[asGoFile] = tgoToGo(tgoFileContents)
				addedGoFiles[asGoFile] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func fillTgoOverlayBasedOnPreviusDriverResponse(dr *DriverResponse, driverOverlay map[string][]byte, addedGoFiles map[string]struct{}) (bool, error) {
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
				asGoFile := tgoToGoExt(tgoFile)
				if _, ok := driverOverlay[asGoFile]; ok {
					continue
				}

				tgoFileContents, err := os.ReadFile(tgoFile)
				if err != nil {
					return false, err
				}

				// TODO: if .go file has the same imports as .tgo, we don't need to rerun.
				// (We need to preserve also the same line and col numbers).

				// TODO: for *all* packages that have tgo, we need to clear ExportData?
				// TODO: what about errors, because of invalid (incomplete) files.
				// Can we detect such case and retry in a go list mode that
				// does not do type checking (only imports)??

				driverOverlay[asGoFile] = tgoToGo(tgoFileContents)
				addedGoFiles[asGoFile] = struct{}{}

				// TODO: in case of the same imports, do not re-run.
				// how this affects the ExportFile clearing?
				// We might not get an error (while loading), but the ExportFile
				// might not contain all decarations that are in the .tgo file.
				// (Tgo file has the same imports as Go file, but Tgo file has one more global func).
				added = true
			}
		}
	}

	return added, nil
}
