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
