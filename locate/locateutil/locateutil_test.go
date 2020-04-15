package locateutil_test

import (
	"regexp"
	"strings"
	"testing"

	"cloudeng.io/go/locate/locateutil"
	"golang.org/x/tools/go/packages"
)

var packagesConfig = &packages.Config{
	Mode: packages.NeedName | packages.NeedSyntax |
		packages.NeedTypes | packages.NeedTypesInfo | packages.NeedCompiledGoFiles,
	Tests:      false,
	BuildFlags: nil,
}

func TestImports(t *testing.T) {
	pkgs, err := packages.Load(packagesConfig,
		"cloudeng.io/go/locate/testdata/imports",
	)
	if err != nil {
		t.Errorf("pkg.Load: %v", err)
	}
	pkg := pkgs[0]
	fns := locateutil.Functions(pkg, regexp.MustCompile(".*"), true)
	start, end := locateutil.ImportBlock(fns[0].File)
	if got, want := pkg.Fset.Position(start).String(), "blocks.go:3:1"; !strings.HasSuffix(got, want) {
		t.Errorf("got %v, doesn't have suffix %v\n", got, want)
	}
	if got, want := pkg.Fset.Position(end).String(), "blocks.go:8:2"; !strings.HasSuffix(got, want) {
		t.Errorf("got %v, doesn't have suffix %v\n", got, want)
	}
}
