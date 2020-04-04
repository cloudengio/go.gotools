package locateutil_test

import (
	"regexp"
	"strings"
	"testing"

	"cloudeng.io/go/locate/locateutil"
	"golang.org/x/tools/go/packages"
)

func TestFunction(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedCompiledGoFiles,
		Tests:      false,
		BuildFlags: nil,
	}
	pkgs, err := packages.Load(cfg,
		"cloudeng.io/go/locate/testdata/data",
		"cloudeng.io/go/locate/testdata/impl",
	)
	if err != nil {
		t.Errorf("pkg.Load: %v", err)
	}
	var fns []locateutil.FuncDesc
	// Top-level functions only.
	for _, pkg := range pkgs {
		fns = append(fns, locateutil.Functions(pkg, regexp.MustCompile(".*"), true)...)
	}
	if got, want := len(fns), 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	// Top-level functions and methods.
	fns = nil
	for _, pkg := range pkgs {
		fns = append(fns, locateutil.Functions(pkg, regexp.MustCompile(".*"), false)...)
	}
	if got, want := len(fns), 16; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	for i, tc := range []struct {
		name, pos string
	}{
		{"Fn1", "data/functions.go:7:6"},
		{"Fn1", "data/functions.go:13:15"},
		{"Fn2", "data/functions_more.go:3:6"},
		{"M1", "interfaces.go:4:2"},
		{"M2", "interfaces.go:5:2"},
		{"M1", "interfaces.go:9:2"},
		{"M3", "interfaces.go:13:2"},
		{"M1", "impl/impls.go:5:17"},
		{"M2", "impl/impls.go:9:17"},
		{"M3", "impl/impls.go:15:17"},
		{"M1", "impl/impls.go:22:18"},
		{"M2", "impl/impls.go:26:18"},
		{"M3", "impl/impls.go:30:18"},
		{"M1", "impl/impls.go:37:17"},
		{"M1", "impl/impls.go:44:3"},
		{"M2", "impl/impls.go:45:3"},
	} {
		if got, want := fns[i].Type.Name(), tc.name; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
		if got, want := fns[i].Position.String(), tc.pos; !strings.HasSuffix(got, want) {
			t.Errorf("%v: got %v doesn't suffix %v", i, got, want)
		}
		if fns[i].File == nil {
			t.Errorf("missing ast for %v", i)
		}
		if fns[i].Decl == nil && !fns[i].Abstract {
			t.Errorf("missing declaration for %v", i)
		}
	}
}
