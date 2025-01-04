package packages_test

import (
	"log"
	"testing"

	"github.com/tgo-lang/tools/go/packages"
	"github.com/tgo-lang/tools/internal/packagestest"
)

func TestTgo(t *testing.T) {
	testAllOrModulesParallel(t, testTgo)
}

func testTgo(t *testing.T, exporter packagestest.Exporter) {
	log.SetFlags(log.Lshortfile)
	exported := packagestest.Export(t, exporter, []packagestest.Module{
		{
			Name: "fake",
			Files: map[string]any{
				"fake.tgo": `package test

import "fake/other"
import "github.com/mateusz834/tgo"

func lool(c tgo.Ctx) error {
	return other.Test(c)
}
`,
				"other/other.tgo": `package other

import "github.com/mateusz834/tgo"

func Test(c tgo.Ctx) error {
	<div>
		"\{1.3}"
	</div>
	return nil
}
`,
			},
		},
		{
			Name: "github.com/mateusz834/tgo",
			Files: map[string]any{
				"tgo.go": `package tgo
type Ctx struct{}
type Error = error
type UnsafeHTML string
type DynamicWriteAllowed interface {string|UnsafeHTML|int|uint|rune}
func DynamicWrite[T DynamicWriteAllowed](t T) {}
`,
			},
		},
	})
	t.Cleanup(exported.Cleanup)

	exported.Config.Mode = packages.LoadSyntax | packages.NeedDeps | packages.NeedExportFile
	pkgs, err := packages.Load(exported.Config, "fake")
	if err != nil {
		t.Fatal(err)
	}

	packages.PrintErrors(pkgs)
}

func TestTgoAndGoInOverlay(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	exported := packagestest.Export(t, packagestest.Modules, []packagestest.Module{
		{
			Name: "test",
			Overlay: map[string][]byte{
				"test.go":  []byte(`package test; func TestOutdated2() {}`),
				"test.tgo": []byte(`package test; func Test2() {}`),
			},
			Files: map[string]any{
				"test.go":  `package test; func TestOutdated() {}`,
				"test.tgo": `package test; func Test() {}`,
			},
		},
	})
	t.Cleanup(exported.Cleanup)

	exported.Config.Mode = packages.LoadSyntax | packages.NeedDeps | packages.NeedExportFile
	pkgs, err := packages.Load(exported.Config, "test")
	if err != nil {
		t.Fatal(err)
	}
	loadExpectNoErrors(t, pkgs)

	if len(pkgs) == 0 || pkgs[0].Name != "test" {
		t.Fatalf("packages.Load() returned unexpected packages: %v", pkgs)
	}

	if pkgs[0].Types.Scope().Lookup("Test2") == nil {
		t.Errorf("test package does not have Test2 symbol defined")
	}

	if l := len(pkgs[0].Types.Scope().Names()); l != 1 {
		t.Errorf("test packages has: %v symbols in Scope, want = 1", l)
	}
}

func TestGoFileInOverlay(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	exported := packagestest.Export(t, packagestest.Modules, []packagestest.Module{
		{
			Name: "test",
			Overlay: map[string][]byte{
				"test.go": []byte(`package test; func TestOutdated2() {}`),
			},
			Files: map[string]any{
				"test.go":  `package test; func TestOutdated() {}`,
				"test.tgo": `package test; func Test() {}`,
			},
		},
	})
	t.Cleanup(exported.Cleanup)

	exported.Config.Mode = packages.LoadSyntax | packages.NeedDeps | packages.NeedExportFile
	pkgs, err := packages.Load(exported.Config, "test")
	if err != nil {
		t.Fatal(err)
	}
	loadExpectNoErrors(t, pkgs)

	if len(pkgs) == 0 || pkgs[0].Name != "test" {
		t.Fatalf("packages.Load() returned unexpected packages: %v", pkgs)
	}

	if pkgs[0].Types.Scope().Lookup("Test") == nil {
		t.Errorf("test package does not have Test symbol defined")
	}

	if l := len(pkgs[0].Types.Scope().Names()); l != 1 {
		t.Errorf("test packages has: %v symbols in Scope, want = 1", l)
	}
}

func loadExpectNoErrors(t *testing.T, pkgs []*packages.Package) {
	t.Helper()
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			t.Errorf("%v: unexpected error: %v", pkg.Name, err)
		}
	})
}
