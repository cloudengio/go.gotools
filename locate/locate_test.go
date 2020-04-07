package locate_test

import (
	"context"
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"
	"testing"

	"cloudeng.io/errors"
	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

const here = "cloudeng.io/go/locate/testdata/"

func listInterfaces(locator *locate.T) []string {
	out := []string{}
	locator.WalkInterfaces(func(name string, pkg *packages.Package,
		file *ast.File, decl *ast.TypeSpec, ifc *types.Interface) {
		line := fmt.Sprintf("%s interface %s", name, pkg.Fset.PositionFor(decl.Pos(), false))
		out = append(out, line)
	})
	return out
}

func listFunctions(locator *locate.T) []string {
	out := []string{}
	locator.WalkFunctions(func(name string, pkg *packages.Package, file *ast.File, fn *types.Func, decl *ast.FuncDecl, implements []string) {
		line := name
		if len(implements) > 0 {
			line += fmt.Sprintf(" implements %s", strings.Join(implements, ", "))
		}
		line += fmt.Sprintf(" @ %s", pkg.Fset.PositionFor(decl.Type.Func, false))
		out = append(out, line)
	})
	return out
}

func listFiles(locator *locate.T) []string {
	out := []string{}
	locator.WalkFiles(func(name string, pkg *packages.Package, comments ast.CommentMap, file *ast.File, hitMask locate.HitMask) {
		if hitMask == 0 {
			return
		}
		line := fmt.Sprintf("%s: %s (%s)", name, file.Name, hitMask)
		out = append(out, line)
	})
	return out
}

func listPackages(locator *locate.T) []string {
	out := []string{}
	locator.WalkPackages(func(pkg *packages.Package) {
		out = append(out, pkg.PkgPath)
	})
	return out
}

func compareLocations(t *testing.T, locations []string, prefixes, suffixes []string) {
	loc := errors.Caller(2, 1)
	if got, want := len(locations), len(suffixes); got != want {
		t.Errorf("%v: got %v, want %v", loc, got, want)
		return
	}
	sort.Strings(locations)
	for i, l := range locations {
		if got, want := l, prefixes[i]; !strings.HasPrefix(got, want) {
			t.Errorf("%v: %v doesn't start with %v", loc, got, want)
		}
		if got, want := l, suffixes[i]; !strings.HasSuffix(got, want) {
			t.Errorf("%v: got %v doesn't have suffix %v", loc, got, want)
		}
	}
}

func compareFiles(t *testing.T, found []string, expected ...string) {
	loc := errors.Caller(2, 1)
	sort.Strings(found)
	for i, f := range found {
		if got, want := f, expected[i]; !strings.Contains(got, want) {
			t.Errorf("%v: got %v doesn't have suffix %v", loc, got, want)
		}
	}
}

func compareSlices(t *testing.T, got, want []string) {
	if got, want := len(got), len(want); got != want {
		t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
		return
	}
	for i := range got {
		if got, want := got[i], want[i]; !strings.HasSuffix(got, want) {
			t.Errorf("%v: got %v does not end with %v", errors.Caller(2, 1), got, want)
			return
		}
	}
}

func TestMultiPackageError(t *testing.T) {
	ctx := context.Background()

	locator := locate.New()
	locator.AddFunctions(here+"data.nomatch", "notapackage")
	err := locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to find: notapackage") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "multipackage")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to type check: cloudeng.io/go/locate/testdata/multipackage") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "parseerror")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to type check: cloudeng.io/go/locate/testdata/parseerror") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "typeerror")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to type check: cloudeng.io/go/locate/testdata/typeerror") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "data.(")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to compile regexp") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}
}

func TestHitMask(t *testing.T) {
	for i, tc := range []struct {
		hm  locate.HitMask
		out string
	}{
		{locate.HasComment, "comment"},
		{locate.HasFunction, "function"},
		{locate.HasInterface, "interface"},
		{locate.HasComment | locate.HasInterface, "comment, interface"},
		{locate.HasInterface | locate.HasComment, "comment, interface"},
	} {
		if got, want := tc.hm.String(), tc.out; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
	}
}
